const fs = require('fs-extra');
const path = require('path');
const ejs = require('ejs');
const collectJSONS = require('./collect-jsons');
const {
  calculatePercentage,
  createReportFolders,
  formatDuration,
  getCustomStyleSheet,
  getGenericJsContent,
  getStyleSheet,
} = require('./utils');
const { parseScenarioSteps } = require('./parse.cucumber.data');

const INDEX_HTML = 'index.html';
const FEATURE_FOLDER = 'features';

function generateReport(options) {
  if (!options) {
    throw new Error('Options need to be provided.');
  }

  if (!options.jsonDir) {
    throw new Error('A path which holds the JSON files should be provided.');
  }

  if (!options.reportPath) {
    throw new Error('An output path for the reports should be defined, no path was provided.');
  }

  const customMetadata = options.customMetadata || false;
  const customData = options.customData || null;
  const style = getStyleSheet(options.overrideStyle) + getCustomStyleSheet(options.customStyle);
  const ingress = options.ingress || { name:'XXXXXXXX', version: 'XXXXXXXX'};
  const reportPath = path.resolve(process.cwd(), options.reportPath);
  const durationInMS = options.durationInMS || false;
  const pageFooter = options.pageFooter || false;
  const buildTime = options.buildTime || "N/A";

  createReportFolders(reportPath);

  const allFeatures = collectJSONS(options);

  const suite = {
    customMetadata,
    customData,
    style,
    name: '',
    version: 'version',
    time: new Date(),
    features: allFeatures,
    ingress,
    totalFeaturesCount: {
      ambiguous: {
        count: 0,
        percentage: 0,
      },
      failed: {
        count: 0,
        percentage: 0,
      },
      passed: {
        count: 0,
        percentage: 0,
      },
      notDefined: {
        count: 0,
        percentage: 0,
      },
      pending: {
        count: 0,
        percentage: 0,
      },
      skipped: {
        count: 0,
        percentage: 0,
      },
      total: 0,
    },
    totalScenariosCount: {
      ambiguous: {
        count: 0,
        percentage: 0,
      },
      failed: {
        count: 0,
        percentage: 0,
      },
      passed: {
        count: 0,
        percentage: 0,
      },
      notDefined: {
        count: 0,
        percentage: 0,
      },
      pending: {
        count: 0,
        percentage: 0,
      },
      skipped: {
        count: 0,
        percentage: 0,
      },
      total: 0,
    },
    totalTime: 0,
  };

  parseFeatures(suite);

  // Percentages
  suite.totalFeaturesCount = calculatePercentage(suite.totalFeaturesCount);

  createFeaturesOverviewIndexPage(suite);
  createFeatureIndexPages(suite);

  // console.log(JSON.stringify(suite, null,2));
  console.log(`\n
=====================================================================================
    Multiple Cucumber HTML report generated in:

    ${path.join(reportPath, INDEX_HTML)}
=====================================================================================\n`);
  // console.log(JSON.stringify(suite, null, 2));

  function parseFeatures(suite) {
    suite.features.forEach((feature) => {
      feature.totalFeatureScenariosCount = {
        ambiguous: {
          count: 0,
          percentage: 0,
        },
        failed: {
          count: 0,
          percentage: 0,
        },
        passed: {
          count: 0,
          percentage: 0,
        },
        notDefined: {
          count: 0,
          percentage: 0,
        },
        pending: {
          count: 0,
          percentage: 0,
        },
        skipped: {
          count: 0,
          percentage: 0,
        },
        total: 0,
      };
      feature.duration = 0;
      feature.time = '0s';
      feature.isFailed = false;
      feature.isAmbiguous = false;
      feature.isSkipped = false;
      feature.isNotdefined = false;
      feature.isPending = false;
      suite.totalFeaturesCount.total++;
      feature.id = `${feature.id}`.replace(/[^a-zA-Z0-9-_]/g, '-');

      if (!feature.elements) {
        return;
      }

      feature = parseScenarios(feature, suite);

      if (feature.isFailed) {
        feature.failed++;
        suite.totalFeaturesCount.failed.count++;
      } else if (feature.isAmbiguous) {
        feature.ambiguous++;
        suite.totalFeaturesCount.ambiguous.count++;
      } else if (feature.isNotdefined) {
        feature.notDefined++;
        suite.totalFeaturesCount.notDefined.count++;
      } else if (feature.isPending) {
        feature.pending++;
        suite.totalFeaturesCount.pending.count++;
      } else if (feature.isSkipped) {
        feature.skipped++;
        suite.totalFeaturesCount.skipped.count++;
      } else {
        feature.passed++;
        suite.totalFeaturesCount.passed.count++;
      }

      if (feature.duration) {
        feature.totalTime += feature.duration;
        feature.time = formatDuration(durationInMS, feature.duration);
      }

      // Percentages
      feature.totalFeatureScenariosCount = calculatePercentage(feature.totalFeatureScenariosCount);
      suite.totalScenariosCount = calculatePercentage(suite.totalScenariosCount);
    });
  }

  /**
     * Parse each scenario within a feature
     * @param {object} feature a feature with all the scenarios in it
     * @return {object} return the parsed feature
     * @private
     */
  function parseScenarios(feature) {
    feature.elements.forEach((scenario) => {
      scenario.passed = 0;
      scenario.failed = 0;
      scenario.notDefined = 0;
      scenario.skipped = 0;
      scenario.pending = 0;
      scenario.ambiguous = 0;
      scenario.duration = 0;
      scenario.time = '0s';
      scenario = parseScenarioSteps(scenario, durationInMS);

      if (scenario.duration > 0) {
        feature.duration += scenario.duration;
        scenario.time = formatDuration(durationInMS, scenario.duration);
      }

      if (scenario.hasOwnProperty('description') && scenario.description) {
        scenario.description = scenario.description.replace(new RegExp('\r?\n', 'g'), '<br />');
      }

      if (scenario.failed > 0) {
        suite.totalScenariosCount.total++;
        suite.totalScenariosCount.failed.count++;
        feature.totalFeatureScenariosCount.total++;
        feature.isFailed = true;

        return feature.totalFeatureScenariosCount.failed.count++;
      }

      if (scenario.ambiguous > 0) {
        suite.totalScenariosCount.total++;
        suite.totalScenariosCount.ambiguous.count++;
        feature.totalFeatureScenariosCount.total++;
        feature.isAmbiguous = true;

        return feature.totalFeatureScenariosCount.ambiguous.count++;
      }

      if (scenario.notDefined > 0) {
        suite.totalScenariosCount.total++;
        suite.totalScenariosCount.notDefined.count++;
        feature.totalFeatureScenariosCount.total++;
        feature.isNotdefined = true;

        return feature.totalFeatureScenariosCount.notDefined.count++;
      }

      if (scenario.pending > 0) {
        suite.totalScenariosCount.total++;
        suite.totalScenariosCount.pending.count++;
        feature.totalFeatureScenariosCount.total++;
        feature.isPending = true;

        return feature.totalFeatureScenariosCount.pending.count++;
      }

      if (scenario.skipped > 0) {
        suite.totalScenariosCount.total++;
        suite.totalScenariosCount.skipped.count++;
        feature.totalFeatureScenariosCount.total++;

        return feature.totalFeatureScenariosCount.skipped.count++;
      }

      /* istanbul ignore else */
      if (scenario.passed > 0) {
        suite.totalScenariosCount.total++;
        suite.totalScenariosCount.passed.count++;
        feature.totalFeatureScenariosCount.total++;

        return feature.totalFeatureScenariosCount.passed.count++;
      }
    });

    feature.isSkipped = feature.totalFeatureScenariosCount.total === feature.totalFeatureScenariosCount.skipped.count;

    return feature;
  }

  function getScenarioStatus(scenario) {
    if (scenario.failed) {
      return 'danger';
    }else if (scenario.ambiguous) {
      return 'warning';
    }else if (scenario.pending) {
      return 'info'
    }else if (scenario.skipped) {
      return 'secondary'
    }else if (scenario.passed) {
      return 'success'
    }
  }

  /**
     * Generate the features overview
     * @param {object} suite JSON object with all the features and scenarios
     * @private
     */
  function createFeaturesOverviewIndexPage(suite) {
    ejs.renderFile(
      path.join(__dirname, '..', 'templates', 'features-overview.index.ejs'),
      {
        ...{ suite },
        ...{
          genericScript: getGenericJsContent(),
          pageFooter,
          buildTime,
          ingress,
          styles: suite.style,
        },
        getScenarioStatus,
      },
      {
        rmWhitespace: true
      },
      (err, str) => {
        if (err) {
          console.log('err = ', err);
          return;
        }

        fs.writeFileSync(
          path.resolve(reportPath, INDEX_HTML),
          str,
        );
      },
    );
  }

  /**
     * Generate the feature pages
     * @param suite suite JSON object with all the features and scenarios
     * @private
     */
  function createFeatureIndexPages(suite) {
    suite.features.forEach((feature) => {
      const featurePage = path.resolve(reportPath, `${FEATURE_FOLDER}/${feature.id}.html`);
      ejs.renderFile(
        path.join(__dirname, '..', 'templates', 'feature-overview.index.ejs'),
        {
          ...{ suite },
          ...{ feature },
          ...{
            genericScript: getGenericJsContent(),
            pageFooter,
            buildTime,
            ingress,
            styles: suite.style,
          },
          getScenarioStatus,
        },
        {},
        (err, str) => {
          if (err) {
            console.log('err = ', err);
            return;
          }

          fs.writeFileSync(
            featurePage,
            str,
          );
        },
      );
    });
  }
}

module.exports = {
  generate: generateReport,
};
