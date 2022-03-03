const { readFileSync, statSync } = require('fs-extra');
const { findJsonFiles, formatToLocalIso } = require('./utils');
const { parseFeatureHooks, parseMetadata } = require('./parse.cucumber.data');

module.exports = function collectJSONS(options) {
  const jsonOutput = [];
  const files = findJsonFiles(options.jsonDir);

  if (files.length === 0) {
    console.log(`WARNING: No JSON files found in '${options.jsonDir}'. NO REPORT CAN BE CREATED!`);
    return [];
  }

  files.map((file) => {
    let data;
    // Cucumber json can be  empty, it's likely being created by another process (#47)
    // or the data could not be a valid JSON-file
    try {
      data = JSON.parse(readFileSync(file).toString());
    } catch (e) {
      data = [];
      console.log(`WARNING: File: '${file}' had no valid JSON data due to error:'${e}'. CONTENT WAS NOT LOADED!`);
    }

    const jsonData = Array.isArray(data) ? data : [data];
    const stats = statSync(file);
    const reportTime = formatToLocalIso(stats.birthtime);

    jsonData.map((json) => {
      json = parseMetadata(json, options.metadata);

      if (options.displayReportTime) {
        json.metadata = {
          ...json.metadata,
          ...{ reportTime },
        };
      }

      // Only check the feature hooks if there are elements (fail safe)
      const { elements } = json;

      if (elements) {
        json.elements = elements.map((scenario) => {
          const { before, after } = scenario;

          if (before) {
            scenario.steps = parseFeatureHooks(before, 'Before')
              .concat(scenario.steps);
          }
          if (after) {
            scenario.steps = scenario.steps
              .concat(parseFeatureHooks(after, 'After'));
          }

          return scenario;
        });
      }

      jsonOutput.push(json);
    });
  });

  return jsonOutput;
};
