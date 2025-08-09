var __defProp = Object.defineProperty;
var __defProps = Object.defineProperties;
var __getOwnPropDescs = Object.getOwnPropertyDescriptors;
var __getOwnPropSymbols = Object.getOwnPropertySymbols;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __propIsEnum = Object.prototype.propertyIsEnumerable;
var __defNormalProp = (obj, key, value) => key in obj ? __defProp(obj, key, { enumerable: true, configurable: true, writable: true, value }) : obj[key] = value;
var __spreadValues = (a, b) => {
  for (var prop in b || (b = {}))
    if (__hasOwnProp.call(b, prop))
      __defNormalProp(a, prop, b[prop]);
  if (__getOwnPropSymbols)
    for (var prop of __getOwnPropSymbols(b)) {
      if (__propIsEnum.call(b, prop))
        __defNormalProp(a, prop, b[prop]);
    }
  return a;
};
var __spreadProps = (a, b) => __defProps(a, __getOwnPropDescs(b));

// generated/api.ts
import globalAxios2 from "axios";

// generated/base.ts
import globalAxios from "axios";
var BASE_PATH = "http://localhost:8080/api/v1".replace(/\/+$/, "");
var BaseAPI = class {
  constructor(configuration, basePath = BASE_PATH, axios = globalAxios) {
    this.basePath = basePath;
    this.axios = axios;
    var _a;
    if (configuration) {
      this.configuration = configuration;
      this.basePath = (_a = configuration.basePath) != null ? _a : basePath;
    }
  }
};
var RequiredError = class extends Error {
  constructor(field, msg) {
    super(msg);
    this.field = field;
    this.name = "RequiredError";
  }
};
var operationServerMap = {};

// generated/common.ts
var DUMMY_BASE_URL = "https://example.com";
var assertParamExists = function(functionName, paramName, paramValue) {
  if (paramValue === null || paramValue === void 0) {
    throw new RequiredError(paramName, `Required parameter ${paramName} was null or undefined when calling ${functionName}.`);
  }
};
var setApiKeyToObject = async function(object, keyParamName, configuration) {
  if (configuration && configuration.apiKey) {
    const localVarApiKeyValue = typeof configuration.apiKey === "function" ? await configuration.apiKey(keyParamName) : await configuration.apiKey;
    object[keyParamName] = localVarApiKeyValue;
  }
};
function setFlattenedQueryParams(urlSearchParams, parameter, key = "") {
  if (parameter == null) return;
  if (typeof parameter === "object") {
    if (Array.isArray(parameter)) {
      parameter.forEach((item) => setFlattenedQueryParams(urlSearchParams, item, key));
    } else {
      Object.keys(parameter).forEach(
        (currentKey) => setFlattenedQueryParams(urlSearchParams, parameter[currentKey], `${key}${key !== "" ? "." : ""}${currentKey}`)
      );
    }
  } else {
    if (urlSearchParams.has(key)) {
      urlSearchParams.append(key, parameter);
    } else {
      urlSearchParams.set(key, parameter);
    }
  }
}
var setSearchParams = function(url, ...objects) {
  const searchParams = new URLSearchParams(url.search);
  setFlattenedQueryParams(searchParams, objects);
  url.search = searchParams.toString();
};
var serializeDataIfNeeded = function(value, requestOptions, configuration) {
  const nonString = typeof value !== "string";
  const needsSerialization = nonString && configuration && configuration.isJsonMime ? configuration.isJsonMime(requestOptions.headers["Content-Type"]) : nonString;
  return needsSerialization ? JSON.stringify(value !== void 0 ? value : {}) : value || "";
};
var toPathString = function(url) {
  return url.pathname + url.search + url.hash;
};
var createRequestFunction = function(axiosArgs, globalAxios3, BASE_PATH2, configuration) {
  return (axios = globalAxios3, basePath = BASE_PATH2) => {
    var _a;
    const axiosRequestArgs = __spreadProps(__spreadValues({}, axiosArgs.options), { url: (axios.defaults.baseURL ? "" : (_a = configuration == null ? void 0 : configuration.basePath) != null ? _a : basePath) + axiosArgs.url });
    return axios.request(axiosRequestArgs);
  };
};

// generated/api.ts
var MediaBatchParamsSortEnum = {
  CreateDate: "createDate"
};
var ApiKeysApiAxiosParamCreator = function(configuration) {
  return {
    /**
     * 
     * @summary Create a new api key
     * @param {ApiKeyParams} params The new token params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createApiKey: async (params, options = {}) => {
      assertParamExists("createApiKey", "params", params);
      const localVarPath = `/keys`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(params, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Delete an api key
     * @param {string} tokenId Api key id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteApiKey: async (tokenId, options = {}) => {
      assertParamExists("deleteApiKey", "tokenId", tokenId);
      const localVarPath = `/keys/{tokenId}`.replace(`{${"tokenId"}}`, encodeURIComponent(String(tokenId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "DELETE" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get all api keys
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getApiKeys: async (options = {}) => {
      const localVarPath = `/keys`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    }
  };
};
var ApiKeysApiFp = function(configuration) {
  const localVarAxiosParamCreator = ApiKeysApiAxiosParamCreator(configuration);
  return {
    /**
     * 
     * @summary Create a new api key
     * @param {ApiKeyParams} params The new token params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async createApiKey(params, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.createApiKey(params, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ApiKeysApi.createApiKey"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Delete an api key
     * @param {string} tokenId Api key id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async deleteApiKey(tokenId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.deleteApiKey(tokenId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ApiKeysApi.deleteApiKey"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get all api keys
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getApiKeys(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getApiKeys(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ApiKeysApi.getApiKeys"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    }
  };
};
var ApiKeysApiFactory = function(configuration, basePath, axios) {
  const localVarFp = ApiKeysApiFp(configuration);
  return {
    /**
     * 
     * @summary Create a new api key
     * @param {ApiKeyParams} params The new token params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createApiKey(params, options) {
      return localVarFp.createApiKey(params, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Delete an api key
     * @param {string} tokenId Api key id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteApiKey(tokenId, options) {
      return localVarFp.deleteApiKey(tokenId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get all api keys
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getApiKeys(options) {
      return localVarFp.getApiKeys(options).then((request) => request(axios, basePath));
    }
  };
};
var ApiKeysApi = class extends BaseAPI {
  /**
   * 
   * @summary Create a new api key
   * @param {ApiKeyParams} params The new token params
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof ApiKeysApi
   */
  createApiKey(params, options) {
    return ApiKeysApiFp(this.configuration).createApiKey(params, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Delete an api key
   * @param {string} tokenId Api key id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof ApiKeysApi
   */
  deleteApiKey(tokenId, options) {
    return ApiKeysApiFp(this.configuration).deleteApiKey(tokenId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get all api keys
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof ApiKeysApi
   */
  getApiKeys(options) {
    return ApiKeysApiFp(this.configuration).getApiKeys(options).then((request) => request(this.axios, this.basePath));
  }
};
var FilesApiAxiosParamCreator = function(configuration) {
  return {
    /**
     * 
     * @summary Add a file to an upload task
     * @param {string} uploadId Upload Id
     * @param {NewFilesParams} request New file params
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    addFilesToUpload: async (uploadId, request, shareId, options = {}) => {
      assertParamExists("addFilesToUpload", "uploadId", uploadId);
      assertParamExists("addFilesToUpload", "request", request);
      const localVarPath = `/upload/{uploadId}`.replace(`{${"uploadId"}}`, encodeURIComponent(String(uploadId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get path completion suggestions
     * @param {string} searchPath Search path
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    autocompletePath: async (searchPath, options = {}) => {
      assertParamExists("autocompletePath", "searchPath", searchPath);
      const localVarPath = `/files/autocomplete`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (searchPath !== void 0) {
        localVarQueryParameter["searchPath"] = searchPath;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * Dispatch a task to create a zip file of the given files, or get the id of a previously created zip file if it already exists
     * @summary Create a zip file
     * @param {FilesListParams} request File Ids
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createTakeout: async (request, shareId, options = {}) => {
      assertParamExists("createTakeout", "request", request);
      const localVarPath = `/takeout`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Delete Files \"permanently\"
     * @param {FilesListParams} request Delete files request body
     * @param {boolean} [ignoreTrash] Delete files even if they are not in the trash
     * @param {boolean} [preserveFolder] Preserve parent folder if it is empty after deletion
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteFiles: async (request, ignoreTrash, preserveFolder, options = {}) => {
      assertParamExists("deleteFiles", "request", request);
      const localVarPath = `/files`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "DELETE" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (ignoreTrash !== void 0) {
        localVarQueryParameter["ignoreTrash"] = ignoreTrash;
      }
      if (preserveFolder !== void 0) {
        localVarQueryParameter["preserveFolder"] = preserveFolder;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Download a file
     * @param {string} fileId File Id
     * @param {string} [shareId] Share Id
     * @param {string} [format] File format conversion
     * @param {boolean} [isTakeout] Is this a takeout file
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    downloadFile: async (fileId, shareId, format, isTakeout, options = {}) => {
      assertParamExists("downloadFile", "fileId", fileId);
      const localVarPath = `/files/{fileId}/download`.replace(`{${"fileId"}}`, encodeURIComponent(String(fileId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      if (format !== void 0) {
        localVarQueryParameter["format"] = format;
      }
      if (isTakeout !== void 0) {
        localVarQueryParameter["isTakeout"] = isTakeout;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get information about a file
     * @param {string} fileId File Id
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFile: async (fileId, shareId, options = {}) => {
      assertParamExists("getFile", "fileId", fileId);
      const localVarPath = `/files/{fileId}`.replace(`{${"fileId"}}`, encodeURIComponent(String(fileId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get the statistics of a file
     * @param {string} fileId File Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFileStats: async (fileId, options = {}) => {
      assertParamExists("getFileStats", "fileId", fileId);
      const localVarPath = `/files/{fileId}/stats`.replace(`{${"fileId"}}`, encodeURIComponent(String(fileId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get the text of a text file
     * @param {string} fileId File Id
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFileText: async (fileId, shareId, options = {}) => {
      assertParamExists("getFileText", "fileId", fileId);
      const localVarPath = `/files/{fileId}/text`.replace(`{${"fileId"}}`, encodeURIComponent(String(fileId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get files shared with the logged in user
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getSharedFiles: async (options = {}) => {
      const localVarPath = `/files/shared`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get the result of an upload task. This will block until the upload is complete
     * @param {string} uploadId Upload Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getUploadResult: async (uploadId, options = {}) => {
      assertParamExists("getUploadResult", "uploadId", uploadId);
      const localVarPath = `/upload/{uploadId}`.replace(`{${"uploadId"}}`, encodeURIComponent(String(uploadId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Move a list of files to a new parent folder
     * @param {MoveFilesParams} request Move files request body
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    moveFiles: async (request, shareId, options = {}) => {
      assertParamExists("moveFiles", "request", request);
      const localVarPath = `/files`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary structsore files from some time in the past
     * @param {RestoreFilesBody} request RestoreFiles files request body
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    restoreFiles: async (request, options = {}) => {
      assertParamExists("restoreFiles", "request", request);
      const localVarPath = `/files/structsore`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Search for files by filename
     * @param {string} search Filename to search for
     * @param {string} [baseFolderId] The folder to search in, defaults to the user\&#39;s home folder
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    searchByFilename: async (search, baseFolderId, options = {}) => {
      assertParamExists("searchByFilename", "search", search);
      const localVarPath = `/files/search`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (search !== void 0) {
        localVarQueryParameter["search"] = search;
      }
      if (baseFolderId !== void 0) {
        localVarQueryParameter["baseFolderId"] = baseFolderId;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Begin a new upload task
     * @param {NewUploadParams} request New upload request body
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    startUpload: async (request, shareId, options = {}) => {
      assertParamExists("startUpload", "request", request);
      const localVarPath = `/upload`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Move a list of files out of the trash, structsoring them to where they were before
     * @param {FilesListParams} request Un-trash files request body
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    unTrashFiles: async (request, options = {}) => {
      assertParamExists("unTrashFiles", "request", request);
      const localVarPath = `/files/untrash`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Update a File
     * @param {string} fileId File Id
     * @param {UpdateFileParams} request Update file request body
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    updateFile: async (fileId, request, shareId, options = {}) => {
      assertParamExists("updateFile", "fileId", fileId);
      assertParamExists("updateFile", "request", request);
      const localVarPath = `/files/{fileId}`.replace(`{${"fileId"}}`, encodeURIComponent(String(fileId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Add a chunk to a file upload
     * @param {string} uploadId Upload Id
     * @param {string} fileId File Id
     * @param {File} chunk File chunk
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    uploadFileChunk: async (uploadId, fileId, chunk, shareId, options = {}) => {
      assertParamExists("uploadFileChunk", "uploadId", uploadId);
      assertParamExists("uploadFileChunk", "fileId", fileId);
      assertParamExists("uploadFileChunk", "chunk", chunk);
      const localVarPath = `/upload/{uploadId}/file/{fileId}`.replace(`{${"uploadId"}}`, encodeURIComponent(String(uploadId))).replace(`{${"fileId"}}`, encodeURIComponent(String(fileId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PUT" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      const localVarFormParams = new (configuration && configuration.formDataCtor || FormData)();
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      if (chunk !== void 0) {
        localVarFormParams.append("chunk", chunk);
      }
      localVarHeaderParameter["Content-Type"] = "multipart/form-data";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = localVarFormParams;
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    }
  };
};
var FilesApiFp = function(configuration) {
  const localVarAxiosParamCreator = FilesApiAxiosParamCreator(configuration);
  return {
    /**
     * 
     * @summary Add a file to an upload task
     * @param {string} uploadId Upload Id
     * @param {NewFilesParams} request New file params
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async addFilesToUpload(uploadId, request, shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.addFilesToUpload(uploadId, request, shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.addFilesToUpload"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get path completion suggestions
     * @param {string} searchPath Search path
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async autocompletePath(searchPath, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.autocompletePath(searchPath, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.autocompletePath"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * Dispatch a task to create a zip file of the given files, or get the id of a previously created zip file if it already exists
     * @summary Create a zip file
     * @param {FilesListParams} request File Ids
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async createTakeout(request, shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.createTakeout(request, shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.createTakeout"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Delete Files \"permanently\"
     * @param {FilesListParams} request Delete files request body
     * @param {boolean} [ignoreTrash] Delete files even if they are not in the trash
     * @param {boolean} [preserveFolder] Preserve parent folder if it is empty after deletion
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async deleteFiles(request, ignoreTrash, preserveFolder, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.deleteFiles(request, ignoreTrash, preserveFolder, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.deleteFiles"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Download a file
     * @param {string} fileId File Id
     * @param {string} [shareId] Share Id
     * @param {string} [format] File format conversion
     * @param {boolean} [isTakeout] Is this a takeout file
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async downloadFile(fileId, shareId, format, isTakeout, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.downloadFile(fileId, shareId, format, isTakeout, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.downloadFile"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get information about a file
     * @param {string} fileId File Id
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFile(fileId, shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFile(fileId, shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.getFile"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get the statistics of a file
     * @param {string} fileId File Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFileStats(fileId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFileStats(fileId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.getFileStats"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get the text of a text file
     * @param {string} fileId File Id
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFileText(fileId, shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFileText(fileId, shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.getFileText"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get files shared with the logged in user
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getSharedFiles(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getSharedFiles(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.getSharedFiles"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get the result of an upload task. This will block until the upload is complete
     * @param {string} uploadId Upload Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getUploadResult(uploadId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getUploadResult(uploadId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.getUploadResult"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Move a list of files to a new parent folder
     * @param {MoveFilesParams} request Move files request body
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async moveFiles(request, shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.moveFiles(request, shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.moveFiles"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary structsore files from some time in the past
     * @param {RestoreFilesBody} request RestoreFiles files request body
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async restoreFiles(request, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.restoreFiles(request, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.restoreFiles"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Search for files by filename
     * @param {string} search Filename to search for
     * @param {string} [baseFolderId] The folder to search in, defaults to the user\&#39;s home folder
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async searchByFilename(search, baseFolderId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.searchByFilename(search, baseFolderId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.searchByFilename"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Begin a new upload task
     * @param {NewUploadParams} request New upload request body
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async startUpload(request, shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.startUpload(request, shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.startUpload"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Move a list of files out of the trash, structsoring them to where they were before
     * @param {FilesListParams} request Un-trash files request body
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async unTrashFiles(request, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.unTrashFiles(request, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.unTrashFiles"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Update a File
     * @param {string} fileId File Id
     * @param {UpdateFileParams} request Update file request body
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async updateFile(fileId, request, shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.updateFile(fileId, request, shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.updateFile"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Add a chunk to a file upload
     * @param {string} uploadId Upload Id
     * @param {string} fileId File Id
     * @param {File} chunk File chunk
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async uploadFileChunk(uploadId, fileId, chunk, shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.uploadFileChunk(uploadId, fileId, chunk, shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.uploadFileChunk"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    }
  };
};
var FilesApiFactory = function(configuration, basePath, axios) {
  const localVarFp = FilesApiFp(configuration);
  return {
    /**
     * 
     * @summary Add a file to an upload task
     * @param {string} uploadId Upload Id
     * @param {NewFilesParams} request New file params
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    addFilesToUpload(uploadId, request, shareId, options) {
      return localVarFp.addFilesToUpload(uploadId, request, shareId, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Get path completion suggestions
     * @param {string} searchPath Search path
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    autocompletePath(searchPath, options) {
      return localVarFp.autocompletePath(searchPath, options).then((request) => request(axios, basePath));
    },
    /**
     * Dispatch a task to create a zip file of the given files, or get the id of a previously created zip file if it already exists
     * @summary Create a zip file
     * @param {FilesListParams} request File Ids
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createTakeout(request, shareId, options) {
      return localVarFp.createTakeout(request, shareId, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Delete Files \"permanently\"
     * @param {FilesListParams} request Delete files request body
     * @param {boolean} [ignoreTrash] Delete files even if they are not in the trash
     * @param {boolean} [preserveFolder] Preserve parent folder if it is empty after deletion
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteFiles(request, ignoreTrash, preserveFolder, options) {
      return localVarFp.deleteFiles(request, ignoreTrash, preserveFolder, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Download a file
     * @param {string} fileId File Id
     * @param {string} [shareId] Share Id
     * @param {string} [format] File format conversion
     * @param {boolean} [isTakeout] Is this a takeout file
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    downloadFile(fileId, shareId, format, isTakeout, options) {
      return localVarFp.downloadFile(fileId, shareId, format, isTakeout, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get information about a file
     * @param {string} fileId File Id
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFile(fileId, shareId, options) {
      return localVarFp.getFile(fileId, shareId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get the statistics of a file
     * @param {string} fileId File Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFileStats(fileId, options) {
      return localVarFp.getFileStats(fileId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get the text of a text file
     * @param {string} fileId File Id
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFileText(fileId, shareId, options) {
      return localVarFp.getFileText(fileId, shareId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get files shared with the logged in user
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getSharedFiles(options) {
      return localVarFp.getSharedFiles(options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get the result of an upload task. This will block until the upload is complete
     * @param {string} uploadId Upload Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getUploadResult(uploadId, options) {
      return localVarFp.getUploadResult(uploadId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Move a list of files to a new parent folder
     * @param {MoveFilesParams} request Move files request body
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    moveFiles(request, shareId, options) {
      return localVarFp.moveFiles(request, shareId, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary structsore files from some time in the past
     * @param {RestoreFilesBody} request RestoreFiles files request body
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    restoreFiles(request, options) {
      return localVarFp.restoreFiles(request, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Search for files by filename
     * @param {string} search Filename to search for
     * @param {string} [baseFolderId] The folder to search in, defaults to the user\&#39;s home folder
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    searchByFilename(search, baseFolderId, options) {
      return localVarFp.searchByFilename(search, baseFolderId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Begin a new upload task
     * @param {NewUploadParams} request New upload request body
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    startUpload(request, shareId, options) {
      return localVarFp.startUpload(request, shareId, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Move a list of files out of the trash, structsoring them to where they were before
     * @param {FilesListParams} request Un-trash files request body
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    unTrashFiles(request, options) {
      return localVarFp.unTrashFiles(request, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Update a File
     * @param {string} fileId File Id
     * @param {UpdateFileParams} request Update file request body
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    updateFile(fileId, request, shareId, options) {
      return localVarFp.updateFile(fileId, request, shareId, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Add a chunk to a file upload
     * @param {string} uploadId Upload Id
     * @param {string} fileId File Id
     * @param {File} chunk File chunk
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    uploadFileChunk(uploadId, fileId, chunk, shareId, options) {
      return localVarFp.uploadFileChunk(uploadId, fileId, chunk, shareId, options).then((request) => request(axios, basePath));
    }
  };
};
var FilesApi = class extends BaseAPI {
  /**
   * 
   * @summary Add a file to an upload task
   * @param {string} uploadId Upload Id
   * @param {NewFilesParams} request New file params
   * @param {string} [shareId] Share Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  addFilesToUpload(uploadId, request, shareId, options) {
    return FilesApiFp(this.configuration).addFilesToUpload(uploadId, request, shareId, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get path completion suggestions
   * @param {string} searchPath Search path
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  autocompletePath(searchPath, options) {
    return FilesApiFp(this.configuration).autocompletePath(searchPath, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * Dispatch a task to create a zip file of the given files, or get the id of a previously created zip file if it already exists
   * @summary Create a zip file
   * @param {FilesListParams} request File Ids
   * @param {string} [shareId] Share Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  createTakeout(request, shareId, options) {
    return FilesApiFp(this.configuration).createTakeout(request, shareId, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Delete Files \"permanently\"
   * @param {FilesListParams} request Delete files request body
   * @param {boolean} [ignoreTrash] Delete files even if they are not in the trash
   * @param {boolean} [preserveFolder] Preserve parent folder if it is empty after deletion
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  deleteFiles(request, ignoreTrash, preserveFolder, options) {
    return FilesApiFp(this.configuration).deleteFiles(request, ignoreTrash, preserveFolder, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Download a file
   * @param {string} fileId File Id
   * @param {string} [shareId] Share Id
   * @param {string} [format] File format conversion
   * @param {boolean} [isTakeout] Is this a takeout file
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  downloadFile(fileId, shareId, format, isTakeout, options) {
    return FilesApiFp(this.configuration).downloadFile(fileId, shareId, format, isTakeout, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get information about a file
   * @param {string} fileId File Id
   * @param {string} [shareId] Share Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  getFile(fileId, shareId, options) {
    return FilesApiFp(this.configuration).getFile(fileId, shareId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get the statistics of a file
   * @param {string} fileId File Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  getFileStats(fileId, options) {
    return FilesApiFp(this.configuration).getFileStats(fileId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get the text of a text file
   * @param {string} fileId File Id
   * @param {string} [shareId] Share Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  getFileText(fileId, shareId, options) {
    return FilesApiFp(this.configuration).getFileText(fileId, shareId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get files shared with the logged in user
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  getSharedFiles(options) {
    return FilesApiFp(this.configuration).getSharedFiles(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get the result of an upload task. This will block until the upload is complete
   * @param {string} uploadId Upload Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  getUploadResult(uploadId, options) {
    return FilesApiFp(this.configuration).getUploadResult(uploadId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Move a list of files to a new parent folder
   * @param {MoveFilesParams} request Move files request body
   * @param {string} [shareId] Share Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  moveFiles(request, shareId, options) {
    return FilesApiFp(this.configuration).moveFiles(request, shareId, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary structsore files from some time in the past
   * @param {RestoreFilesBody} request RestoreFiles files request body
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  restoreFiles(request, options) {
    return FilesApiFp(this.configuration).restoreFiles(request, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Search for files by filename
   * @param {string} search Filename to search for
   * @param {string} [baseFolderId] The folder to search in, defaults to the user\&#39;s home folder
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  searchByFilename(search, baseFolderId, options) {
    return FilesApiFp(this.configuration).searchByFilename(search, baseFolderId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Begin a new upload task
   * @param {NewUploadParams} request New upload request body
   * @param {string} [shareId] Share Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  startUpload(request, shareId, options) {
    return FilesApiFp(this.configuration).startUpload(request, shareId, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Move a list of files out of the trash, structsoring them to where they were before
   * @param {FilesListParams} request Un-trash files request body
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  unTrashFiles(request, options) {
    return FilesApiFp(this.configuration).unTrashFiles(request, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Update a File
   * @param {string} fileId File Id
   * @param {UpdateFileParams} request Update file request body
   * @param {string} [shareId] Share Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  updateFile(fileId, request, shareId, options) {
    return FilesApiFp(this.configuration).updateFile(fileId, request, shareId, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Add a chunk to a file upload
   * @param {string} uploadId Upload Id
   * @param {string} fileId File Id
   * @param {File} chunk File chunk
   * @param {string} [shareId] Share Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FilesApi
   */
  uploadFileChunk(uploadId, fileId, chunk, shareId, options) {
    return FilesApiFp(this.configuration).uploadFileChunk(uploadId, fileId, chunk, shareId, options).then((request) => request(this.axios, this.basePath));
  }
};
var FolderApiAxiosParamCreator = function(configuration) {
  return {
    /**
     * 
     * @summary Create a new folder
     * @param {CreateFolderBody} request New folder body
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createFolder: async (request, shareId, options = {}) => {
      assertParamExists("createFolder", "request", request);
      const localVarPath = `/folder`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get a folder
     * @param {string} folderId Folder Id
     * @param {string} [shareId] Share Id
     * @param {number} [timestamp] Past timestamp to view the folder at, in ms since epoch
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFolder: async (folderId, shareId, timestamp, options = {}) => {
      assertParamExists("getFolder", "folderId", folderId);
      const localVarPath = `/folder/{folderId}`.replace(`{${"folderId"}}`, encodeURIComponent(String(folderId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      if (timestamp !== void 0) {
        localVarQueryParameter["timestamp"] = timestamp;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get actions of a folder at a given time
     * @param {string} fileId File Id
     * @param {number} timestamp Past timestamp to view the folder at, in ms since epoch
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFolderHistory: async (fileId, timestamp, options = {}) => {
      assertParamExists("getFolderHistory", "fileId", fileId);
      assertParamExists("getFolderHistory", "timestamp", timestamp);
      const localVarPath = `/files/{fileId}/history`.replace(`{${"fileId"}}`, encodeURIComponent(String(fileId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (timestamp !== void 0) {
        localVarQueryParameter["timestamp"] = timestamp;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Dispatch a folder scan
     * @param {StructsScanBody} request Scan parameters
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    scanFolder: async (request, shareId, options = {}) => {
      assertParamExists("scanFolder", "request", request);
      const localVarPath = `/folder/scan`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Set the cover image of a folder
     * @param {string} folderId Folder Id
     * @param {string} mediaId Media Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setFolderCover: async (folderId, mediaId, options = {}) => {
      assertParamExists("setFolderCover", "folderId", folderId);
      assertParamExists("setFolderCover", "mediaId", mediaId);
      const localVarPath = `/folder/{folderId}/cover`.replace(`{${"folderId"}}`, encodeURIComponent(String(folderId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (mediaId !== void 0) {
        localVarQueryParameter["mediaId"] = mediaId;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    }
  };
};
var FolderApiFp = function(configuration) {
  const localVarAxiosParamCreator = FolderApiAxiosParamCreator(configuration);
  return {
    /**
     * 
     * @summary Create a new folder
     * @param {CreateFolderBody} request New folder body
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async createFolder(request, shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.createFolder(request, shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FolderApi.createFolder"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get a folder
     * @param {string} folderId Folder Id
     * @param {string} [shareId] Share Id
     * @param {number} [timestamp] Past timestamp to view the folder at, in ms since epoch
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFolder(folderId, shareId, timestamp, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFolder(folderId, shareId, timestamp, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FolderApi.getFolder"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get actions of a folder at a given time
     * @param {string} fileId File Id
     * @param {number} timestamp Past timestamp to view the folder at, in ms since epoch
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFolderHistory(fileId, timestamp, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFolderHistory(fileId, timestamp, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FolderApi.getFolderHistory"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Dispatch a folder scan
     * @param {StructsScanBody} request Scan parameters
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async scanFolder(request, shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.scanFolder(request, shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FolderApi.scanFolder"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Set the cover image of a folder
     * @param {string} folderId Folder Id
     * @param {string} mediaId Media Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async setFolderCover(folderId, mediaId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.setFolderCover(folderId, mediaId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FolderApi.setFolderCover"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    }
  };
};
var FolderApiFactory = function(configuration, basePath, axios) {
  const localVarFp = FolderApiFp(configuration);
  return {
    /**
     * 
     * @summary Create a new folder
     * @param {CreateFolderBody} request New folder body
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createFolder(request, shareId, options) {
      return localVarFp.createFolder(request, shareId, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Get a folder
     * @param {string} folderId Folder Id
     * @param {string} [shareId] Share Id
     * @param {number} [timestamp] Past timestamp to view the folder at, in ms since epoch
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFolder(folderId, shareId, timestamp, options) {
      return localVarFp.getFolder(folderId, shareId, timestamp, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get actions of a folder at a given time
     * @param {string} fileId File Id
     * @param {number} timestamp Past timestamp to view the folder at, in ms since epoch
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFolderHistory(fileId, timestamp, options) {
      return localVarFp.getFolderHistory(fileId, timestamp, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Dispatch a folder scan
     * @param {StructsScanBody} request Scan parameters
     * @param {string} [shareId] Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    scanFolder(request, shareId, options) {
      return localVarFp.scanFolder(request, shareId, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Set the cover image of a folder
     * @param {string} folderId Folder Id
     * @param {string} mediaId Media Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setFolderCover(folderId, mediaId, options) {
      return localVarFp.setFolderCover(folderId, mediaId, options).then((request) => request(axios, basePath));
    }
  };
};
var FolderApi = class extends BaseAPI {
  /**
   * 
   * @summary Create a new folder
   * @param {CreateFolderBody} request New folder body
   * @param {string} [shareId] Share Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FolderApi
   */
  createFolder(request, shareId, options) {
    return FolderApiFp(this.configuration).createFolder(request, shareId, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get a folder
   * @param {string} folderId Folder Id
   * @param {string} [shareId] Share Id
   * @param {number} [timestamp] Past timestamp to view the folder at, in ms since epoch
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FolderApi
   */
  getFolder(folderId, shareId, timestamp, options) {
    return FolderApiFp(this.configuration).getFolder(folderId, shareId, timestamp, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get actions of a folder at a given time
   * @param {string} fileId File Id
   * @param {number} timestamp Past timestamp to view the folder at, in ms since epoch
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FolderApi
   */
  getFolderHistory(fileId, timestamp, options) {
    return FolderApiFp(this.configuration).getFolderHistory(fileId, timestamp, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Dispatch a folder scan
   * @param {StructsScanBody} request Scan parameters
   * @param {string} [shareId] Share Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FolderApi
   */
  scanFolder(request, shareId, options) {
    return FolderApiFp(this.configuration).scanFolder(request, shareId, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Set the cover image of a folder
   * @param {string} folderId Folder Id
   * @param {string} mediaId Media Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof FolderApi
   */
  setFolderCover(folderId, mediaId, options) {
    return FolderApiFp(this.configuration).setFolderCover(folderId, mediaId, options).then((request) => request(this.axios, this.basePath));
  }
};
var MediaApiAxiosParamCreator = function(configuration) {
  return {
    /**
     * 
     * @summary Make sure all media is correctly synced with the file system
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    cleanupMedia: async (options = {}) => {
      const localVarPath = `/media/cleanup`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Drop all computed media HDIR data. Must be server owner.
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    dropHDIRs: async (options = {}) => {
      const localVarPath = `/media/drop/hdirs`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary DANGEROUS. Drop all computed media and clear thumbnail in-memory and filesystem cache. Must be server owner.
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    dropMedia: async (options = {}) => {
      const localVarPath = `/media/drop`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get paginated media
     * @param {MediaBatchParams} request Media Batch Params
     * @param {string} [shareId] File ShareId
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMedia: async (request, shareId, options = {}) => {
      assertParamExists("getMedia", "request", request);
      const localVarPath = `/media`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get file of media by id
     * @param {string} mediaId Id of media
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaFile: async (mediaId, options = {}) => {
      assertParamExists("getMediaFile", "mediaId", mediaId);
      const localVarPath = `/media/{mediaId}/file`.replace(`{${"mediaId"}}`, encodeURIComponent(String(mediaId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get a media image bytes
     * @param {string} mediaId Media Id
     * @param {string} extension Extension
     * @param {GetMediaImageQualityEnum} quality Image Quality
     * @param {number} [page] Page number
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaImage: async (mediaId, extension, quality, page, options = {}) => {
      assertParamExists("getMediaImage", "mediaId", mediaId);
      assertParamExists("getMediaImage", "extension", extension);
      assertParamExists("getMediaImage", "quality", quality);
      const localVarPath = `/media/{mediaId}.{extension}`.replace(`{${"mediaId"}}`, encodeURIComponent(String(mediaId))).replace(`{${"extension"}}`, encodeURIComponent(String(extension)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (quality !== void 0) {
        localVarQueryParameter["quality"] = quality;
      }
      if (page !== void 0) {
        localVarQueryParameter["page"] = page;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get media info
     * @param {string} mediaId Media Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaInfo: async (mediaId, options = {}) => {
      assertParamExists("getMediaInfo", "mediaId", mediaId);
      const localVarPath = `/media/{mediaId}/info`.replace(`{${"mediaId"}}`, encodeURIComponent(String(mediaId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get media type dictionary
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaTypes: async (options = {}) => {
      const localVarPath = `/media/types`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get random media
     * @param {number} count Number of random medias to get
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getRandomMedia: async (count, options = {}) => {
      assertParamExists("getRandomMedia", "count", count);
      const localVarPath = `/media/random`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (count !== void 0) {
        localVarQueryParameter["count"] = count;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Like a media
     * @param {string} mediaId Id of media
     * @param {boolean} liked Liked status to set
     * @param {string} [shareId] ShareId
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setMediaLiked: async (mediaId, liked, shareId, options = {}) => {
      assertParamExists("setMediaLiked", "mediaId", mediaId);
      assertParamExists("setMediaLiked", "liked", liked);
      const localVarPath = `/media/{mediaId}/liked`.replace(`{${"mediaId"}}`, encodeURIComponent(String(mediaId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareId !== void 0) {
        localVarQueryParameter["shareId"] = shareId;
      }
      if (liked !== void 0) {
        localVarQueryParameter["liked"] = liked;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Set media visibility
     * @param {boolean} hidden Set the media visibility
     * @param {MediaIdsParams} mediaIds MediaIds to change visibility of
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setMediaVisibility: async (hidden, mediaIds, options = {}) => {
      assertParamExists("setMediaVisibility", "hidden", hidden);
      assertParamExists("setMediaVisibility", "mediaIds", mediaIds);
      const localVarPath = `/media/visibility`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (hidden !== void 0) {
        localVarQueryParameter["hidden"] = hidden;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(mediaIds, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Stream a video
     * @param {string} mediaId Id of media
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    streamVideo: async (mediaId, options = {}) => {
      assertParamExists("streamVideo", "mediaId", mediaId);
      const localVarPath = `/media/{mediaId}/video`.replace(`{${"mediaId"}}`, encodeURIComponent(String(mediaId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    }
  };
};
var MediaApiFp = function(configuration) {
  const localVarAxiosParamCreator = MediaApiAxiosParamCreator(configuration);
  return {
    /**
     * 
     * @summary Make sure all media is correctly synced with the file system
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async cleanupMedia(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.cleanupMedia(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.cleanupMedia"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Drop all computed media HDIR data. Must be server owner.
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async dropHDIRs(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.dropHDIRs(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.dropHDIRs"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary DANGEROUS. Drop all computed media and clear thumbnail in-memory and filesystem cache. Must be server owner.
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async dropMedia(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.dropMedia(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.dropMedia"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get paginated media
     * @param {MediaBatchParams} request Media Batch Params
     * @param {string} [shareId] File ShareId
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getMedia(request, shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getMedia(request, shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.getMedia"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get file of media by id
     * @param {string} mediaId Id of media
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getMediaFile(mediaId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getMediaFile(mediaId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.getMediaFile"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get a media image bytes
     * @param {string} mediaId Media Id
     * @param {string} extension Extension
     * @param {GetMediaImageQualityEnum} quality Image Quality
     * @param {number} [page] Page number
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getMediaImage(mediaId, extension, quality, page, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getMediaImage(mediaId, extension, quality, page, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.getMediaImage"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get media info
     * @param {string} mediaId Media Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getMediaInfo(mediaId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getMediaInfo(mediaId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.getMediaInfo"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get media type dictionary
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getMediaTypes(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getMediaTypes(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.getMediaTypes"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get random media
     * @param {number} count Number of random medias to get
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getRandomMedia(count, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getRandomMedia(count, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.getRandomMedia"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Like a media
     * @param {string} mediaId Id of media
     * @param {boolean} liked Liked status to set
     * @param {string} [shareId] ShareId
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async setMediaLiked(mediaId, liked, shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.setMediaLiked(mediaId, liked, shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.setMediaLiked"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Set media visibility
     * @param {boolean} hidden Set the media visibility
     * @param {MediaIdsParams} mediaIds MediaIds to change visibility of
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async setMediaVisibility(hidden, mediaIds, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.setMediaVisibility(hidden, mediaIds, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.setMediaVisibility"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Stream a video
     * @param {string} mediaId Id of media
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async streamVideo(mediaId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.streamVideo(mediaId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.streamVideo"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    }
  };
};
var MediaApiFactory = function(configuration, basePath, axios) {
  const localVarFp = MediaApiFp(configuration);
  return {
    /**
     * 
     * @summary Make sure all media is correctly synced with the file system
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    cleanupMedia(options) {
      return localVarFp.cleanupMedia(options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Drop all computed media HDIR data. Must be server owner.
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    dropHDIRs(options) {
      return localVarFp.dropHDIRs(options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary DANGEROUS. Drop all computed media and clear thumbnail in-memory and filesystem cache. Must be server owner.
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    dropMedia(options) {
      return localVarFp.dropMedia(options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get paginated media
     * @param {MediaBatchParams} request Media Batch Params
     * @param {string} [shareId] File ShareId
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMedia(request, shareId, options) {
      return localVarFp.getMedia(request, shareId, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Get file of media by id
     * @param {string} mediaId Id of media
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaFile(mediaId, options) {
      return localVarFp.getMediaFile(mediaId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get a media image bytes
     * @param {string} mediaId Media Id
     * @param {string} extension Extension
     * @param {GetMediaImageQualityEnum} quality Image Quality
     * @param {number} [page] Page number
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaImage(mediaId, extension, quality, page, options) {
      return localVarFp.getMediaImage(mediaId, extension, quality, page, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get media info
     * @param {string} mediaId Media Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaInfo(mediaId, options) {
      return localVarFp.getMediaInfo(mediaId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get media type dictionary
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaTypes(options) {
      return localVarFp.getMediaTypes(options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get random media
     * @param {number} count Number of random medias to get
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getRandomMedia(count, options) {
      return localVarFp.getRandomMedia(count, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Like a media
     * @param {string} mediaId Id of media
     * @param {boolean} liked Liked status to set
     * @param {string} [shareId] ShareId
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setMediaLiked(mediaId, liked, shareId, options) {
      return localVarFp.setMediaLiked(mediaId, liked, shareId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Set media visibility
     * @param {boolean} hidden Set the media visibility
     * @param {MediaIdsParams} mediaIds MediaIds to change visibility of
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setMediaVisibility(hidden, mediaIds, options) {
      return localVarFp.setMediaVisibility(hidden, mediaIds, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Stream a video
     * @param {string} mediaId Id of media
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    streamVideo(mediaId, options) {
      return localVarFp.streamVideo(mediaId, options).then((request) => request(axios, basePath));
    }
  };
};
var MediaApi = class extends BaseAPI {
  /**
   * 
   * @summary Make sure all media is correctly synced with the file system
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof MediaApi
   */
  cleanupMedia(options) {
    return MediaApiFp(this.configuration).cleanupMedia(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Drop all computed media HDIR data. Must be server owner.
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof MediaApi
   */
  dropHDIRs(options) {
    return MediaApiFp(this.configuration).dropHDIRs(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary DANGEROUS. Drop all computed media and clear thumbnail in-memory and filesystem cache. Must be server owner.
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof MediaApi
   */
  dropMedia(options) {
    return MediaApiFp(this.configuration).dropMedia(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get paginated media
   * @param {MediaBatchParams} request Media Batch Params
   * @param {string} [shareId] File ShareId
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof MediaApi
   */
  getMedia(request, shareId, options) {
    return MediaApiFp(this.configuration).getMedia(request, shareId, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get file of media by id
   * @param {string} mediaId Id of media
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof MediaApi
   */
  getMediaFile(mediaId, options) {
    return MediaApiFp(this.configuration).getMediaFile(mediaId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get a media image bytes
   * @param {string} mediaId Media Id
   * @param {string} extension Extension
   * @param {GetMediaImageQualityEnum} quality Image Quality
   * @param {number} [page] Page number
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof MediaApi
   */
  getMediaImage(mediaId, extension, quality, page, options) {
    return MediaApiFp(this.configuration).getMediaImage(mediaId, extension, quality, page, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get media info
   * @param {string} mediaId Media Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof MediaApi
   */
  getMediaInfo(mediaId, options) {
    return MediaApiFp(this.configuration).getMediaInfo(mediaId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get media type dictionary
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof MediaApi
   */
  getMediaTypes(options) {
    return MediaApiFp(this.configuration).getMediaTypes(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get random media
   * @param {number} count Number of random medias to get
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof MediaApi
   */
  getRandomMedia(count, options) {
    return MediaApiFp(this.configuration).getRandomMedia(count, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Like a media
   * @param {string} mediaId Id of media
   * @param {boolean} liked Liked status to set
   * @param {string} [shareId] ShareId
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof MediaApi
   */
  setMediaLiked(mediaId, liked, shareId, options) {
    return MediaApiFp(this.configuration).setMediaLiked(mediaId, liked, shareId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Set media visibility
   * @param {boolean} hidden Set the media visibility
   * @param {MediaIdsParams} mediaIds MediaIds to change visibility of
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof MediaApi
   */
  setMediaVisibility(hidden, mediaIds, options) {
    return MediaApiFp(this.configuration).setMediaVisibility(hidden, mediaIds, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Stream a video
   * @param {string} mediaId Id of media
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof MediaApi
   */
  streamVideo(mediaId, options) {
    return MediaApiFp(this.configuration).streamVideo(mediaId, options).then((request) => request(this.axios, this.basePath));
  }
};
var GetMediaImageQualityEnum = {
  Thumbnail: "thumbnail",
  Fullres: "fullres"
};
var ShareApiAxiosParamCreator = function(configuration) {
  return {
    /**
     * 
     * @summary Add a user to a file share
     * @param {string} shareId Share Id
     * @param {AddUserParams} request Share Accessors
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    addUserToShare: async (shareId, request, options = {}) => {
      assertParamExists("addUserToShare", "shareId", shareId);
      assertParamExists("addUserToShare", "request", request);
      const localVarPath = `/share/{shareId}/accessors`.replace(`{${"shareId"}}`, encodeURIComponent(String(shareId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Share a file
     * @param {FileShareParams} request New File Share Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createFileShare: async (request, options = {}) => {
      assertParamExists("createFileShare", "request", request);
      const localVarPath = `/share/file`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Delete a file share
     * @param {string} shareId Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteFileShare: async (shareId, options = {}) => {
      assertParamExists("deleteFileShare", "shareId", shareId);
      const localVarPath = `/share/{shareId}`.replace(`{${"shareId"}}`, encodeURIComponent(String(shareId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "DELETE" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get a file share
     * @param {string} shareId Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFileShare: async (shareId, options = {}) => {
      assertParamExists("getFileShare", "shareId", shareId);
      const localVarPath = `/share/{shareId}`.replace(`{${"shareId"}}`, encodeURIComponent(String(shareId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Remove a user from a file share
     * @param {string} shareId Share Id
     * @param {string} username Username
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    removeUserFromShare: async (shareId, username, options = {}) => {
      assertParamExists("removeUserFromShare", "shareId", shareId);
      assertParamExists("removeUserFromShare", "username", username);
      const localVarPath = `/share/{shareId}/accessors/{username}`.replace(`{${"shareId"}}`, encodeURIComponent(String(shareId))).replace(`{${"username"}}`, encodeURIComponent(String(username)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "DELETE" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Update a share\'s \"public\" status
     * @param {string} shareId Share Id
     * @param {boolean} _public Share Public Status
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setSharePublic: async (shareId, _public, options = {}) => {
      assertParamExists("setSharePublic", "shareId", shareId);
      assertParamExists("setSharePublic", "_public", _public);
      const localVarPath = `/share/{shareId}/public`.replace(`{${"shareId"}}`, encodeURIComponent(String(shareId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (_public !== void 0) {
        localVarQueryParameter["public"] = _public;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Update a share\'s user permissions
     * @param {string} shareId Share Id
     * @param {string} username Username
     * @param {PermissionsParams} request Share Permissions Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    updateShareAccessorPermissions: async (shareId, username, request, options = {}) => {
      assertParamExists("updateShareAccessorPermissions", "shareId", shareId);
      assertParamExists("updateShareAccessorPermissions", "username", username);
      assertParamExists("updateShareAccessorPermissions", "request", request);
      const localVarPath = `/share/{shareId}/accessors/{username}`.replace(`{${"shareId"}}`, encodeURIComponent(String(shareId))).replace(`{${"username"}}`, encodeURIComponent(String(username)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    }
  };
};
var ShareApiFp = function(configuration) {
  const localVarAxiosParamCreator = ShareApiAxiosParamCreator(configuration);
  return {
    /**
     * 
     * @summary Add a user to a file share
     * @param {string} shareId Share Id
     * @param {AddUserParams} request Share Accessors
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async addUserToShare(shareId, request, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.addUserToShare(shareId, request, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ShareApi.addUserToShare"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Share a file
     * @param {FileShareParams} request New File Share Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async createFileShare(request, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.createFileShare(request, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ShareApi.createFileShare"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Delete a file share
     * @param {string} shareId Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async deleteFileShare(shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.deleteFileShare(shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ShareApi.deleteFileShare"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get a file share
     * @param {string} shareId Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFileShare(shareId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFileShare(shareId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ShareApi.getFileShare"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Remove a user from a file share
     * @param {string} shareId Share Id
     * @param {string} username Username
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async removeUserFromShare(shareId, username, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.removeUserFromShare(shareId, username, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ShareApi.removeUserFromShare"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Update a share\'s \"public\" status
     * @param {string} shareId Share Id
     * @param {boolean} _public Share Public Status
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async setSharePublic(shareId, _public, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.setSharePublic(shareId, _public, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ShareApi.setSharePublic"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Update a share\'s user permissions
     * @param {string} shareId Share Id
     * @param {string} username Username
     * @param {PermissionsParams} request Share Permissions Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async updateShareAccessorPermissions(shareId, username, request, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.updateShareAccessorPermissions(shareId, username, request, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ShareApi.updateShareAccessorPermissions"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    }
  };
};
var ShareApiFactory = function(configuration, basePath, axios) {
  const localVarFp = ShareApiFp(configuration);
  return {
    /**
     * 
     * @summary Add a user to a file share
     * @param {string} shareId Share Id
     * @param {AddUserParams} request Share Accessors
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    addUserToShare(shareId, request, options) {
      return localVarFp.addUserToShare(shareId, request, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Share a file
     * @param {FileShareParams} request New File Share Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createFileShare(request, options) {
      return localVarFp.createFileShare(request, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Delete a file share
     * @param {string} shareId Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteFileShare(shareId, options) {
      return localVarFp.deleteFileShare(shareId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get a file share
     * @param {string} shareId Share Id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFileShare(shareId, options) {
      return localVarFp.getFileShare(shareId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Remove a user from a file share
     * @param {string} shareId Share Id
     * @param {string} username Username
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    removeUserFromShare(shareId, username, options) {
      return localVarFp.removeUserFromShare(shareId, username, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Update a share\'s \"public\" status
     * @param {string} shareId Share Id
     * @param {boolean} _public Share Public Status
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setSharePublic(shareId, _public, options) {
      return localVarFp.setSharePublic(shareId, _public, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Update a share\'s user permissions
     * @param {string} shareId Share Id
     * @param {string} username Username
     * @param {PermissionsParams} request Share Permissions Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    updateShareAccessorPermissions(shareId, username, request, options) {
      return localVarFp.updateShareAccessorPermissions(shareId, username, request, options).then((request2) => request2(axios, basePath));
    }
  };
};
var ShareApi = class extends BaseAPI {
  /**
   * 
   * @summary Add a user to a file share
   * @param {string} shareId Share Id
   * @param {AddUserParams} request Share Accessors
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof ShareApi
   */
  addUserToShare(shareId, request, options) {
    return ShareApiFp(this.configuration).addUserToShare(shareId, request, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Share a file
   * @param {FileShareParams} request New File Share Params
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof ShareApi
   */
  createFileShare(request, options) {
    return ShareApiFp(this.configuration).createFileShare(request, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Delete a file share
   * @param {string} shareId Share Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof ShareApi
   */
  deleteFileShare(shareId, options) {
    return ShareApiFp(this.configuration).deleteFileShare(shareId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get a file share
   * @param {string} shareId Share Id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof ShareApi
   */
  getFileShare(shareId, options) {
    return ShareApiFp(this.configuration).getFileShare(shareId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Remove a user from a file share
   * @param {string} shareId Share Id
   * @param {string} username Username
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof ShareApi
   */
  removeUserFromShare(shareId, username, options) {
    return ShareApiFp(this.configuration).removeUserFromShare(shareId, username, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Update a share\'s \"public\" status
   * @param {string} shareId Share Id
   * @param {boolean} _public Share Public Status
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof ShareApi
   */
  setSharePublic(shareId, _public, options) {
    return ShareApiFp(this.configuration).setSharePublic(shareId, _public, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Update a share\'s user permissions
   * @param {string} shareId Share Id
   * @param {string} username Username
   * @param {PermissionsParams} request Share Permissions Params
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof ShareApi
   */
  updateShareAccessorPermissions(shareId, username, request, options) {
    return ShareApiFp(this.configuration).updateShareAccessorPermissions(shareId, username, request, options).then((request2) => request2(this.axios, this.basePath));
  }
};
var TowersApiAxiosParamCreator = function(configuration) {
  return {
    /**
     * 
     * @summary Create a new remote
     * @param {NewServerParams} request New Server Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createRemote: async (request, options = {}) => {
      assertParamExists("createRemote", "request", request);
      const localVarPath = `/tower/remote`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Delete a remote
     * @param {string} serverId Server Id to delete
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteRemote: async (serverId, options = {}) => {
      assertParamExists("deleteRemote", "serverId", serverId);
      const localVarPath = `/tower/{serverId}`.replace(`{${"serverId"}}`, encodeURIComponent(String(serverId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "DELETE" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get information about a file
     * @param {string} timestamp Timestamp in milliseconds since epoch
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getBackupInfo: async (timestamp, options = {}) => {
      assertParamExists("getBackupInfo", "timestamp", timestamp);
      const localVarPath = `/tower/backup`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      if (timestamp !== void 0) {
        localVarQueryParameter["timestamp"] = timestamp;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get all remotes
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getRemotes: async (options = {}) => {
      const localVarPath = `/tower`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get server info
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getServerInfo: async (options = {}) => {
      const localVarPath = `/info`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Initialize the target server
     * @param {StructsInitServerParams} request Server initialization body
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    initializeTower: async (request, options = {}) => {
      assertParamExists("initializeTower", "request", request);
      const localVarPath = `/tower/init`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(request, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Launch backup on a tower
     * @param {string} serverId Server ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    launchBackup: async (serverId, options = {}) => {
      assertParamExists("launchBackup", "serverId", serverId);
      const localVarPath = `/tower/{serverId}/backup`.replace(`{${"serverId"}}`, encodeURIComponent(String(serverId)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Reset tower
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    resetTower: async (options = {}) => {
      const localVarPath = `/tower/reset`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    }
  };
};
var TowersApiFp = function(configuration) {
  const localVarAxiosParamCreator = TowersApiAxiosParamCreator(configuration);
  return {
    /**
     * 
     * @summary Create a new remote
     * @param {NewServerParams} request New Server Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async createRemote(request, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.createRemote(request, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.createRemote"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Delete a remote
     * @param {string} serverId Server Id to delete
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async deleteRemote(serverId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.deleteRemote(serverId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.deleteRemote"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get information about a file
     * @param {string} timestamp Timestamp in milliseconds since epoch
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getBackupInfo(timestamp, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getBackupInfo(timestamp, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.getBackupInfo"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get all remotes
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getRemotes(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getRemotes(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.getRemotes"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get server info
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getServerInfo(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getServerInfo(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.getServerInfo"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Initialize the target server
     * @param {StructsInitServerParams} request Server initialization body
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async initializeTower(request, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.initializeTower(request, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.initializeTower"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Launch backup on a tower
     * @param {string} serverId Server ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async launchBackup(serverId, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.launchBackup(serverId, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.launchBackup"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Reset tower
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async resetTower(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.resetTower(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.resetTower"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    }
  };
};
var TowersApiFactory = function(configuration, basePath, axios) {
  const localVarFp = TowersApiFp(configuration);
  return {
    /**
     * 
     * @summary Create a new remote
     * @param {NewServerParams} request New Server Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createRemote(request, options) {
      return localVarFp.createRemote(request, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Delete a remote
     * @param {string} serverId Server Id to delete
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteRemote(serverId, options) {
      return localVarFp.deleteRemote(serverId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get information about a file
     * @param {string} timestamp Timestamp in milliseconds since epoch
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getBackupInfo(timestamp, options) {
      return localVarFp.getBackupInfo(timestamp, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get all remotes
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getRemotes(options) {
      return localVarFp.getRemotes(options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get server info
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getServerInfo(options) {
      return localVarFp.getServerInfo(options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Initialize the target server
     * @param {StructsInitServerParams} request Server initialization body
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    initializeTower(request, options) {
      return localVarFp.initializeTower(request, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Launch backup on a tower
     * @param {string} serverId Server ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    launchBackup(serverId, options) {
      return localVarFp.launchBackup(serverId, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Reset tower
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    resetTower(options) {
      return localVarFp.resetTower(options).then((request) => request(axios, basePath));
    }
  };
};
var TowersApi = class extends BaseAPI {
  /**
   * 
   * @summary Create a new remote
   * @param {NewServerParams} request New Server Params
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof TowersApi
   */
  createRemote(request, options) {
    return TowersApiFp(this.configuration).createRemote(request, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Delete a remote
   * @param {string} serverId Server Id to delete
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof TowersApi
   */
  deleteRemote(serverId, options) {
    return TowersApiFp(this.configuration).deleteRemote(serverId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get information about a file
   * @param {string} timestamp Timestamp in milliseconds since epoch
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof TowersApi
   */
  getBackupInfo(timestamp, options) {
    return TowersApiFp(this.configuration).getBackupInfo(timestamp, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get all remotes
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof TowersApi
   */
  getRemotes(options) {
    return TowersApiFp(this.configuration).getRemotes(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get server info
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof TowersApi
   */
  getServerInfo(options) {
    return TowersApiFp(this.configuration).getServerInfo(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Initialize the target server
   * @param {StructsInitServerParams} request Server initialization body
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof TowersApi
   */
  initializeTower(request, options) {
    return TowersApiFp(this.configuration).initializeTower(request, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Launch backup on a tower
   * @param {string} serverId Server ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof TowersApi
   */
  launchBackup(serverId, options) {
    return TowersApiFp(this.configuration).launchBackup(serverId, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Reset tower
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof TowersApi
   */
  resetTower(options) {
    return TowersApiFp(this.configuration).resetTower(options).then((request) => request(this.axios, this.basePath));
  }
};
var UsersApiAxiosParamCreator = function(configuration) {
  return {
    /**
     * 
     * @summary Update active status of user
     * @param {string} username Username of user to update
     * @param {boolean} setActive Target activation status
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    activateUser: async (username, setActive, options = {}) => {
      assertParamExists("activateUser", "username", username);
      assertParamExists("activateUser", "setActive", setActive);
      const localVarPath = `/users/{username}/active`.replace(`{${"username"}}`, encodeURIComponent(String(username)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      if (setActive !== void 0) {
        localVarQueryParameter["setActive"] = setActive;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Update display name of a user
     * @param {string} username Username of user to update
     * @param {string} newFullName New full name of user
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    changeDisplayName: async (username, newFullName, options = {}) => {
      assertParamExists("changeDisplayName", "username", username);
      assertParamExists("changeDisplayName", "newFullName", newFullName);
      const localVarPath = `/users/{username}/fullName`.replace(`{${"username"}}`, encodeURIComponent(String(username)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      if (newFullName !== void 0) {
        localVarQueryParameter["newFullName"] = newFullName;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Check if username is already taken
     * @param {string} username Username of user to check
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    checkExists: async (username, options = {}) => {
      assertParamExists("checkExists", "username", username);
      const localVarPath = `/users/{username}`.replace(`{${"username"}}`, encodeURIComponent(String(username)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "HEAD" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Create a new user
     * @param {NewUserParams} newUserParams New user params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createUser: async (newUserParams, options = {}) => {
      assertParamExists("createUser", "newUserParams", newUserParams);
      const localVarPath = `/users`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(newUserParams, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Delete a user
     * @param {string} username Username of user to delete
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteUser: async (username, options = {}) => {
      assertParamExists("deleteUser", "username", username);
      const localVarPath = `/users/{username}`.replace(`{${"username"}}`, encodeURIComponent(String(username)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "DELETE" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Gets the user based on the auth token
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getUser: async (options = {}) => {
      const localVarPath = `/users/me`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Get all users, including (possibly) sensitive information like password hashes
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getUsers: async (options = {}) => {
      const localVarPath = `/users`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Login User
     * @param {LoginBody} loginParams Login params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    loginUser: async (loginParams, options = {}) => {
      assertParamExists("loginUser", "loginParams", loginParams);
      const localVarPath = `/users/auth`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(loginParams, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Logout User
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    logoutUser: async (options = {}) => {
      const localVarPath = `/users/logout`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Search for users by username
     * @param {string} search Partial username to search for
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    searchUsers: async (search, options = {}) => {
      assertParamExists("searchUsers", "search", search);
      const localVarPath = `/users/search`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (search !== void 0) {
        localVarQueryParameter["search"] = search;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Update admin status of user
     * @param {string} username Username of user to update
     * @param {boolean} setAdmin Target admin status
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setUserAdmin: async (username, setAdmin, options = {}) => {
      assertParamExists("setUserAdmin", "username", username);
      assertParamExists("setUserAdmin", "setAdmin", setAdmin);
      const localVarPath = `/users/{username}/admin`.replace(`{${"username"}}`, encodeURIComponent(String(username)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      if (setAdmin !== void 0) {
        localVarQueryParameter["setAdmin"] = setAdmin;
      }
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Update user password
     * @param {string} username Username of user to update
     * @param {PasswordUpdateParams} passwordUpdateParams Password update params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    updateUserPassword: async (username, passwordUpdateParams, options = {}) => {
      assertParamExists("updateUserPassword", "username", username);
      assertParamExists("updateUserPassword", "passwordUpdateParams", passwordUpdateParams);
      const localVarPath = `/users/{username}/password`.replace(`{${"username"}}`, encodeURIComponent(String(username)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      await setApiKeyToObject(localVarHeaderParameter, "Authorization", configuration);
      localVarHeaderParameter["Content-Type"] = "application/json";
      setSearchParams(localVarUrlObj, localVarQueryParameter);
      let headersFromBaseOptions = baseOptions && baseOptions.headers ? baseOptions.headers : {};
      localVarRequestOptions.headers = __spreadValues(__spreadValues(__spreadValues({}, localVarHeaderParameter), headersFromBaseOptions), options.headers);
      localVarRequestOptions.data = serializeDataIfNeeded(passwordUpdateParams, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    }
  };
};
var UsersApiFp = function(configuration) {
  const localVarAxiosParamCreator = UsersApiAxiosParamCreator(configuration);
  return {
    /**
     * 
     * @summary Update active status of user
     * @param {string} username Username of user to update
     * @param {boolean} setActive Target activation status
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async activateUser(username, setActive, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.activateUser(username, setActive, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["UsersApi.activateUser"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Update display name of a user
     * @param {string} username Username of user to update
     * @param {string} newFullName New full name of user
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async changeDisplayName(username, newFullName, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.changeDisplayName(username, newFullName, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["UsersApi.changeDisplayName"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Check if username is already taken
     * @param {string} username Username of user to check
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async checkExists(username, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.checkExists(username, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["UsersApi.checkExists"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Create a new user
     * @param {NewUserParams} newUserParams New user params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async createUser(newUserParams, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.createUser(newUserParams, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["UsersApi.createUser"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Delete a user
     * @param {string} username Username of user to delete
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async deleteUser(username, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.deleteUser(username, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["UsersApi.deleteUser"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Gets the user based on the auth token
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getUser(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getUser(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["UsersApi.getUser"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get all users, including (possibly) sensitive information like password hashes
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getUsers(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getUsers(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["UsersApi.getUsers"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Login User
     * @param {LoginBody} loginParams Login params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async loginUser(loginParams, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.loginUser(loginParams, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["UsersApi.loginUser"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Logout User
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async logoutUser(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.logoutUser(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["UsersApi.logoutUser"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Search for users by username
     * @param {string} search Partial username to search for
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async searchUsers(search, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.searchUsers(search, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["UsersApi.searchUsers"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Update admin status of user
     * @param {string} username Username of user to update
     * @param {boolean} setAdmin Target admin status
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async setUserAdmin(username, setAdmin, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.setUserAdmin(username, setAdmin, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["UsersApi.setUserAdmin"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Update user password
     * @param {string} username Username of user to update
     * @param {PasswordUpdateParams} passwordUpdateParams Password update params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async updateUserPassword(username, passwordUpdateParams, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.updateUserPassword(username, passwordUpdateParams, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["UsersApi.updateUserPassword"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    }
  };
};
var UsersApiFactory = function(configuration, basePath, axios) {
  const localVarFp = UsersApiFp(configuration);
  return {
    /**
     * 
     * @summary Update active status of user
     * @param {string} username Username of user to update
     * @param {boolean} setActive Target activation status
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    activateUser(username, setActive, options) {
      return localVarFp.activateUser(username, setActive, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Update display name of a user
     * @param {string} username Username of user to update
     * @param {string} newFullName New full name of user
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    changeDisplayName(username, newFullName, options) {
      return localVarFp.changeDisplayName(username, newFullName, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Check if username is already taken
     * @param {string} username Username of user to check
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    checkExists(username, options) {
      return localVarFp.checkExists(username, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Create a new user
     * @param {NewUserParams} newUserParams New user params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createUser(newUserParams, options) {
      return localVarFp.createUser(newUserParams, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Delete a user
     * @param {string} username Username of user to delete
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteUser(username, options) {
      return localVarFp.deleteUser(username, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Gets the user based on the auth token
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getUser(options) {
      return localVarFp.getUser(options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get all users, including (possibly) sensitive information like password hashes
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getUsers(options) {
      return localVarFp.getUsers(options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Login User
     * @param {LoginBody} loginParams Login params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    loginUser(loginParams, options) {
      return localVarFp.loginUser(loginParams, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Logout User
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    logoutUser(options) {
      return localVarFp.logoutUser(options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Search for users by username
     * @param {string} search Partial username to search for
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    searchUsers(search, options) {
      return localVarFp.searchUsers(search, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Update admin status of user
     * @param {string} username Username of user to update
     * @param {boolean} setAdmin Target admin status
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setUserAdmin(username, setAdmin, options) {
      return localVarFp.setUserAdmin(username, setAdmin, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Update user password
     * @param {string} username Username of user to update
     * @param {PasswordUpdateParams} passwordUpdateParams Password update params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    updateUserPassword(username, passwordUpdateParams, options) {
      return localVarFp.updateUserPassword(username, passwordUpdateParams, options).then((request) => request(axios, basePath));
    }
  };
};
var UsersApi = class extends BaseAPI {
  /**
   * 
   * @summary Update active status of user
   * @param {string} username Username of user to update
   * @param {boolean} setActive Target activation status
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof UsersApi
   */
  activateUser(username, setActive, options) {
    return UsersApiFp(this.configuration).activateUser(username, setActive, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Update display name of a user
   * @param {string} username Username of user to update
   * @param {string} newFullName New full name of user
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof UsersApi
   */
  changeDisplayName(username, newFullName, options) {
    return UsersApiFp(this.configuration).changeDisplayName(username, newFullName, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Check if username is already taken
   * @param {string} username Username of user to check
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof UsersApi
   */
  checkExists(username, options) {
    return UsersApiFp(this.configuration).checkExists(username, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Create a new user
   * @param {NewUserParams} newUserParams New user params
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof UsersApi
   */
  createUser(newUserParams, options) {
    return UsersApiFp(this.configuration).createUser(newUserParams, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Delete a user
   * @param {string} username Username of user to delete
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof UsersApi
   */
  deleteUser(username, options) {
    return UsersApiFp(this.configuration).deleteUser(username, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Gets the user based on the auth token
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof UsersApi
   */
  getUser(options) {
    return UsersApiFp(this.configuration).getUser(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get all users, including (possibly) sensitive information like password hashes
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof UsersApi
   */
  getUsers(options) {
    return UsersApiFp(this.configuration).getUsers(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Login User
   * @param {LoginBody} loginParams Login params
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof UsersApi
   */
  loginUser(loginParams, options) {
    return UsersApiFp(this.configuration).loginUser(loginParams, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Logout User
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof UsersApi
   */
  logoutUser(options) {
    return UsersApiFp(this.configuration).logoutUser(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Search for users by username
   * @param {string} search Partial username to search for
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof UsersApi
   */
  searchUsers(search, options) {
    return UsersApiFp(this.configuration).searchUsers(search, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Update admin status of user
   * @param {string} username Username of user to update
   * @param {boolean} setAdmin Target admin status
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof UsersApi
   */
  setUserAdmin(username, setAdmin, options) {
    return UsersApiFp(this.configuration).setUserAdmin(username, setAdmin, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Update user password
   * @param {string} username Username of user to update
   * @param {PasswordUpdateParams} passwordUpdateParams Password update params
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   * @memberof UsersApi
   */
  updateUserPassword(username, passwordUpdateParams, options) {
    return UsersApiFp(this.configuration).updateUserPassword(username, passwordUpdateParams, options).then((request) => request(this.axios, this.basePath));
  }
};

// AllApi.ts
function WeblensApiFactory(apiEndpoint) {
  return {
    MediaApi: MediaApiFactory({}, apiEndpoint),
    FilesApi: FilesApiFactory({}, apiEndpoint),
    FoldersApi: FolderApiFactory({}, apiEndpoint),
    TowersApi: TowersApiFactory({}, apiEndpoint),
    SharesApi: ShareApiFactory({}, apiEndpoint),
    UsersApi: UsersApiFactory({}, apiEndpoint)
  };
}
export {
  ApiKeysApi,
  ApiKeysApiAxiosParamCreator,
  ApiKeysApiFactory,
  ApiKeysApiFp,
  FilesApi,
  FilesApiAxiosParamCreator,
  FilesApiFactory,
  FilesApiFp,
  FolderApi,
  FolderApiAxiosParamCreator,
  FolderApiFactory,
  FolderApiFp,
  GetMediaImageQualityEnum,
  MediaApi,
  MediaApiAxiosParamCreator,
  MediaApiFactory,
  MediaApiFp,
  MediaBatchParamsSortEnum,
  ShareApi,
  ShareApiAxiosParamCreator,
  ShareApiFactory,
  ShareApiFp,
  TowersApi,
  TowersApiAxiosParamCreator,
  TowersApiFactory,
  TowersApiFp,
  UsersApi,
  UsersApiAxiosParamCreator,
  UsersApiFactory,
  UsersApiFp,
  WeblensApiFactory
};
