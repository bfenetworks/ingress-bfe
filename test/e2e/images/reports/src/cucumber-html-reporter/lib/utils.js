const { join, resolve } = require('path');
const {
  accessSync, constants, ensureDirSync, readdirSync, readFileSync,
} = require('fs-extra');
const moment = require('moment');

/**
 * Find all JSON Files
 *
 * @param {string} dir
 *
 * @returns {string[]}
 */
function findJsonFiles(dir) {
  const folder = resolve(process.cwd(), dir);

  try {
    return readdirSync(folder)
      .filter((file) => file.slice(-5) === '.json')
      .map((file) => join(folder, file));
  } catch (e) {
    throw new Error(`There were issues reading JSON-files from '${folder}'.`);
  }
}

/**
 * Format input date to YYYY/MM/DD HH:mm:ss
 *
 * @param {Date} date
 *
 * @returns {string} formatted date in ISO format local time
 */
function formatToLocalIso(date) {
  return moment(date).format('YYYY/MM/DD HH:mm:ss');
}

/**
 * Create the report folders
 *
 * @param {string} folder
 */
function createReportFolders(folder) {
  ensureDirSync(folder);
  ensureDirSync(resolve(folder, 'features'));
}

/**
 * Format the duration to HH:mm:ss.SSS
 *
 * @param {boolean} durationInMS
 * @param {number} ns
 *
 * @return {string}
 */
function formatDuration(durationInMS, ns) {
  // `moment.utc(#)` needs ms, we now use device by 1000000 to calculate ns to ms
  ft = moment.utc(durationInMS ? ns : ns / 1000000).format('m[m]s[s]');
  if (ft.startsWith("0m")) {
    return ft.substr(2, ft.length);
  }

  return ft;
}

/**
 * Escape html in data string
 *
 * @param {*} data
 *
 * @return {*}
 */
function escapeHtml(data) {
  return (typeof data === 'string' || data instanceof String)
    ? data.replace(/</g, '&lt;').replace(/>/g, '&gt;')
    : data;
}

/**
 * Check if the string a base64 string
 *
 * @param {string} string
 *
 * @return {boolean}
 */
function isBase64(string) {
  const notBase64 = /[^A-Z0-9+\/=]/i;
  const stringLength = string.length;

  if (!stringLength || stringLength % 4 !== 0 || notBase64.test(string)) {
    return false;
  }

  const firstPaddingChar = string.indexOf('=');

  return firstPaddingChar === -1
        || firstPaddingChar === stringLength - 1
        || (firstPaddingChar === stringLength - 2 && string[stringLength - 1] === '=');
}

/**
 * Read a file and return it's content
 *
 * @param {string} fileName
 *
 * @return {*} Content of the file
 */
function getTemplateFileContent(fileName) {
  return readFileSync(join(__dirname, '..', 'templates', fileName), 'utf-8');
}

/**
 * Get the generic JS file content
 *
 * @returns {string}
 */
function getGenericJsContent() {
  return getTemplateFileContent('generic.js');
}

/**
 * Get the custom style sheet content
 *
 * @param {string} fileName
 *
 * @returns {string}
 */
function getCustomStyleSheet(fileName = '') {
  if (fileName) {
    try {
      // This is for getting the content of custom CSS files
      accessSync(fileName, constants.R_OK);

      return readFileSync(fileName, 'utf-8');
    } catch (err) {
      console.log(`WARNING: Custom stylesheet: '${fileName}' could not be loaded due to '${err}'.`);
    }
  }

  return '';
}

/**
 * Get the style sheet content
 *
 * @param {string} fileName
 *
 * @returns {string}
 */
function getStyleSheet(fileName = '') {
  if (fileName) {
    try {
      // This is for getting the content of custom CSS files
      accessSync(fileName, constants.R_OK);

      return readFileSync(fileName, 'utf-8');
    } catch (err) {
      console.log(`WARNING: Override stylesheet: '${fileName}' could not be loaded due to '${err}'. The default will be loaded.`);
    }
  }

  return getTemplateFileContent('style.css');
}

/**
 * Calculate the percentage of keys
 *
 * @example:
 *  {
 *      ambiguous: {
 *          count: 0,
 *          percentage: 0
 *      },
 *      failed: {
 *          count: 0,
 *          percentage: 0
 *      },
 *      passed: {
 *          count: 0,
 *          percentage: 0
 *      },
 *      total: 0,
 *  }
 *
 * @param {object} obj
 *
 * @returns {object}
 */
function calculatePercentage(obj) {
  Object.keys(obj).forEach((key) => {
    if (key !== 'total') {
      obj[key].percentage = ((obj[key].count / obj.total) * 100).toFixed(2);
    }
  });

  return obj;
}

module.exports = {
  calculatePercentage,
  createReportFolders,
  escapeHtml,
  findJsonFiles,
  formatDuration,
  formatToLocalIso,
  getCustomStyleSheet,
  getGenericJsContent,
  getStyleSheet,
  isBase64,
};
