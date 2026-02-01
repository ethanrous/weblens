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
var APIKeysApiAxiosParamCreator = function(configuration) {
  return {
    /**
     * 
     * @summary Create a new api key
     * @param {APIKeyParams} params The new token params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createAPIKey: async (params, options = {}) => {
      assertParamExists("createAPIKey", "params", params);
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
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} tokenID API key id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteAPIKey: async (tokenID, options = {}) => {
      assertParamExists("deleteAPIKey", "tokenID", tokenID);
      const localVarPath = `/keys/{tokenID}`.replace(`{${"tokenID"}}`, encodeURIComponent(String(tokenID)));
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
    getAPIKeys: async (options = {}) => {
      const localVarPath = `/keys`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Accept"] = "application/json";
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
var APIKeysApiFp = function(configuration) {
  const localVarAxiosParamCreator = APIKeysApiAxiosParamCreator(configuration);
  return {
    /**
     * 
     * @summary Create a new api key
     * @param {APIKeyParams} params The new token params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async createAPIKey(params, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.createAPIKey(params, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["APIKeysApi.createAPIKey"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Delete an api key
     * @param {string} tokenID API key id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async deleteAPIKey(tokenID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.deleteAPIKey(tokenID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["APIKeysApi.deleteAPIKey"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get all api keys
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getAPIKeys(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getAPIKeys(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["APIKeysApi.getAPIKeys"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    }
  };
};
var APIKeysApiFactory = function(configuration, basePath, axios) {
  const localVarFp = APIKeysApiFp(configuration);
  return {
    /**
     * 
     * @summary Create a new api key
     * @param {APIKeyParams} params The new token params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createAPIKey(params, options) {
      return localVarFp.createAPIKey(params, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Delete an api key
     * @param {string} tokenID API key id
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteAPIKey(tokenID, options) {
      return localVarFp.deleteAPIKey(tokenID, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get all api keys
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getAPIKeys(options) {
      return localVarFp.getAPIKeys(options).then((request) => request(axios, basePath));
    }
  };
};
var APIKeysApi = class extends BaseAPI {
  /**
   * 
   * @summary Create a new api key
   * @param {APIKeyParams} params The new token params
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  createAPIKey(params, options) {
    return APIKeysApiFp(this.configuration).createAPIKey(params, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Delete an api key
   * @param {string} tokenID API key id
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  deleteAPIKey(tokenID, options) {
    return APIKeysApiFp(this.configuration).deleteAPIKey(tokenID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get all api keys
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getAPIKeys(options) {
    return APIKeysApiFp(this.configuration).getAPIKeys(options).then((request) => request(this.axios, this.basePath));
  }
};
var FeatureFlagsApiAxiosParamCreator = function(configuration) {
  return {
    /**
     * 
     * @summary Get Feature Flags
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFlags: async (options = {}) => {
      const localVarPath = `/flags`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @summary Set Feature Flags
     * @param {Array<StructsSetConfigParam>} request Feature Flag Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setFlags: async (request, options = {}) => {
      assertParamExists("setFlags", "request", request);
      const localVarPath = `/flags`;
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
    }
  };
};
var FeatureFlagsApiFp = function(configuration) {
  const localVarAxiosParamCreator = FeatureFlagsApiAxiosParamCreator(configuration);
  return {
    /**
     * 
     * @summary Get Feature Flags
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFlags(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFlags(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FeatureFlagsApi.getFlags"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Set Feature Flags
     * @param {Array<StructsSetConfigParam>} request Feature Flag Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async setFlags(request, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.setFlags(request, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FeatureFlagsApi.setFlags"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    }
  };
};
var FeatureFlagsApiFactory = function(configuration, basePath, axios) {
  const localVarFp = FeatureFlagsApiFp(configuration);
  return {
    /**
     * 
     * @summary Get Feature Flags
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFlags(options) {
      return localVarFp.getFlags(options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Set Feature Flags
     * @param {Array<StructsSetConfigParam>} request Feature Flag Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setFlags(request, options) {
      return localVarFp.setFlags(request, options).then((request2) => request2(axios, basePath));
    }
  };
};
var FeatureFlagsApi = class extends BaseAPI {
  /**
   * 
   * @summary Get Feature Flags
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getFlags(options) {
    return FeatureFlagsApiFp(this.configuration).getFlags(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Set Feature Flags
   * @param {Array<StructsSetConfigParam>} request Feature Flag Params
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  setFlags(request, options) {
    return FeatureFlagsApiFp(this.configuration).setFlags(request, options).then((request2) => request2(this.axios, this.basePath));
  }
};
var FilesApiAxiosParamCreator = function(configuration) {
  return {
    /**
     * 
     * @summary Add a file to an upload task
     * @param {string} uploadID Upload ID
     * @param {NewFilesParams} request New file params
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    addFilesToUpload: async (uploadID, request, shareID, options = {}) => {
      assertParamExists("addFilesToUpload", "uploadID", uploadID);
      assertParamExists("addFilesToUpload", "request", request);
      const localVarPath = `/upload/{uploadID}`.replace(`{${"uploadID"}}`, encodeURIComponent(String(uploadID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      localVarHeaderParameter["Accept"] = "*/*";
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
      localVarHeaderParameter["Accept"] = "*/*";
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
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createTakeout: async (request, shareID, options = {}) => {
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
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      localVarHeaderParameter["Accept"] = "*/*";
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
     * @param {string} fileID File ID
     * @param {string} [shareID] Share ID
     * @param {string} [format] File format conversion
     * @param {boolean} [isTakeout] Is this a takeout file
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    downloadFile: async (fileID, shareID, format, isTakeout, options = {}) => {
      assertParamExists("downloadFile", "fileID", fileID);
      const localVarPath = `/files/{fileID}/download`.replace(`{${"fileID"}}`, encodeURIComponent(String(fileID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
      }
      if (format !== void 0) {
        localVarQueryParameter["format"] = format;
      }
      if (isTakeout !== void 0) {
        localVarQueryParameter["isTakeout"] = isTakeout;
      }
      localVarHeaderParameter["Accept"] = "application/octet-stream";
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
     * @param {string} fileID File ID
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFile: async (fileID, shareID, options = {}) => {
      assertParamExists("getFile", "fileID", fileID);
      const localVarPath = `/files/{fileID}`.replace(`{${"fileID"}}`, encodeURIComponent(String(fileID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
      }
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} fileID File ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFileStats: async (fileID, options = {}) => {
      assertParamExists("getFileStats", "fileID", fileID);
      const localVarPath = `/files/{fileID}/stats`.replace(`{${"fileID"}}`, encodeURIComponent(String(fileID)));
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
     * @param {string} fileID File ID
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFileText: async (fileID, shareID, options = {}) => {
      assertParamExists("getFileText", "fileID", fileID);
      const localVarPath = `/files/{fileID}/text`.replace(`{${"fileID"}}`, encodeURIComponent(String(fileID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
      }
      localVarHeaderParameter["Accept"] = "text/plain";
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
      localVarHeaderParameter["Accept"] = "*/*";
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
     * @param {string} uploadID Upload ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getUploadResult: async (uploadID, options = {}) => {
      assertParamExists("getUploadResult", "uploadID", uploadID);
      const localVarPath = `/upload/{uploadID}`.replace(`{${"uploadID"}}`, encodeURIComponent(String(uploadID)));
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
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    moveFiles: async (request, shareID, options = {}) => {
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
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
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
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} [baseFolderID] The folder to search in, defaults to the user\&#39;s home folder
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    searchByFilename: async (search, baseFolderID, options = {}) => {
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
      if (baseFolderID !== void 0) {
        localVarQueryParameter["baseFolderID"] = baseFolderID;
      }
      localVarHeaderParameter["Accept"] = "*/*";
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
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    startUpload: async (request, shareID, options = {}) => {
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
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      localVarHeaderParameter["Accept"] = "*/*";
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
     * @param {string} fileID File ID
     * @param {UpdateFileParams} request Update file request body
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    updateFile: async (fileID, request, shareID, options = {}) => {
      assertParamExists("updateFile", "fileID", fileID);
      assertParamExists("updateFile", "request", request);
      const localVarPath = `/files/{fileID}`.replace(`{${"fileID"}}`, encodeURIComponent(String(fileID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
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
     * @param {string} uploadID Upload ID
     * @param {string} fileID File ID
     * @param {File} chunk File chunk
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    uploadFileChunk: async (uploadID, fileID, chunk, shareID, options = {}) => {
      assertParamExists("uploadFileChunk", "uploadID", uploadID);
      assertParamExists("uploadFileChunk", "fileID", fileID);
      assertParamExists("uploadFileChunk", "chunk", chunk);
      const localVarPath = `/upload/{uploadID}/file/{fileID}`.replace(`{${"uploadID"}}`, encodeURIComponent(String(uploadID))).replace(`{${"fileID"}}`, encodeURIComponent(String(fileID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PUT" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      const localVarFormParams = new (configuration && configuration.formDataCtor || FormData)();
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
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
     * @param {string} uploadID Upload ID
     * @param {NewFilesParams} request New file params
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async addFilesToUpload(uploadID, request, shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.addFilesToUpload(uploadID, request, shareID, options);
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
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async createTakeout(request, shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.createTakeout(request, shareID, options);
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
     * @param {string} fileID File ID
     * @param {string} [shareID] Share ID
     * @param {string} [format] File format conversion
     * @param {boolean} [isTakeout] Is this a takeout file
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async downloadFile(fileID, shareID, format, isTakeout, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.downloadFile(fileID, shareID, format, isTakeout, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.downloadFile"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get information about a file
     * @param {string} fileID File ID
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFile(fileID, shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFile(fileID, shareID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.getFile"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get the statistics of a file
     * @param {string} fileID File ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFileStats(fileID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFileStats(fileID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.getFileStats"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get the text of a text file
     * @param {string} fileID File ID
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFileText(fileID, shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFileText(fileID, shareID, options);
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
     * @param {string} uploadID Upload ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getUploadResult(uploadID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getUploadResult(uploadID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.getUploadResult"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Move a list of files to a new parent folder
     * @param {MoveFilesParams} request Move files request body
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async moveFiles(request, shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.moveFiles(request, shareID, options);
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
     * @param {string} [baseFolderID] The folder to search in, defaults to the user\&#39;s home folder
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async searchByFilename(search, baseFolderID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.searchByFilename(search, baseFolderID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.searchByFilename"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Begin a new upload task
     * @param {NewUploadParams} request New upload request body
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async startUpload(request, shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.startUpload(request, shareID, options);
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
     * @param {string} fileID File ID
     * @param {UpdateFileParams} request Update file request body
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async updateFile(fileID, request, shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.updateFile(fileID, request, shareID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FilesApi.updateFile"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Add a chunk to a file upload
     * @param {string} uploadID Upload ID
     * @param {string} fileID File ID
     * @param {File} chunk File chunk
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async uploadFileChunk(uploadID, fileID, chunk, shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.uploadFileChunk(uploadID, fileID, chunk, shareID, options);
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
     * @param {string} uploadID Upload ID
     * @param {NewFilesParams} request New file params
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    addFilesToUpload(uploadID, request, shareID, options) {
      return localVarFp.addFilesToUpload(uploadID, request, shareID, options).then((request2) => request2(axios, basePath));
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
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createTakeout(request, shareID, options) {
      return localVarFp.createTakeout(request, shareID, options).then((request2) => request2(axios, basePath));
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
     * @param {string} fileID File ID
     * @param {string} [shareID] Share ID
     * @param {string} [format] File format conversion
     * @param {boolean} [isTakeout] Is this a takeout file
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    downloadFile(fileID, shareID, format, isTakeout, options) {
      return localVarFp.downloadFile(fileID, shareID, format, isTakeout, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get information about a file
     * @param {string} fileID File ID
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFile(fileID, shareID, options) {
      return localVarFp.getFile(fileID, shareID, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get the statistics of a file
     * @param {string} fileID File ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFileStats(fileID, options) {
      return localVarFp.getFileStats(fileID, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get the text of a text file
     * @param {string} fileID File ID
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFileText(fileID, shareID, options) {
      return localVarFp.getFileText(fileID, shareID, options).then((request) => request(axios, basePath));
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
     * @param {string} uploadID Upload ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getUploadResult(uploadID, options) {
      return localVarFp.getUploadResult(uploadID, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Move a list of files to a new parent folder
     * @param {MoveFilesParams} request Move files request body
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    moveFiles(request, shareID, options) {
      return localVarFp.moveFiles(request, shareID, options).then((request2) => request2(axios, basePath));
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
     * @param {string} [baseFolderID] The folder to search in, defaults to the user\&#39;s home folder
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    searchByFilename(search, baseFolderID, options) {
      return localVarFp.searchByFilename(search, baseFolderID, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Begin a new upload task
     * @param {NewUploadParams} request New upload request body
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    startUpload(request, shareID, options) {
      return localVarFp.startUpload(request, shareID, options).then((request2) => request2(axios, basePath));
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
     * @param {string} fileID File ID
     * @param {UpdateFileParams} request Update file request body
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    updateFile(fileID, request, shareID, options) {
      return localVarFp.updateFile(fileID, request, shareID, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Add a chunk to a file upload
     * @param {string} uploadID Upload ID
     * @param {string} fileID File ID
     * @param {File} chunk File chunk
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    uploadFileChunk(uploadID, fileID, chunk, shareID, options) {
      return localVarFp.uploadFileChunk(uploadID, fileID, chunk, shareID, options).then((request) => request(axios, basePath));
    }
  };
};
var FilesApi = class extends BaseAPI {
  /**
   * 
   * @summary Add a file to an upload task
   * @param {string} uploadID Upload ID
   * @param {NewFilesParams} request New file params
   * @param {string} [shareID] Share ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  addFilesToUpload(uploadID, request, shareID, options) {
    return FilesApiFp(this.configuration).addFilesToUpload(uploadID, request, shareID, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get path completion suggestions
   * @param {string} searchPath Search path
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  autocompletePath(searchPath, options) {
    return FilesApiFp(this.configuration).autocompletePath(searchPath, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * Dispatch a task to create a zip file of the given files, or get the id of a previously created zip file if it already exists
   * @summary Create a zip file
   * @param {FilesListParams} request File Ids
   * @param {string} [shareID] Share ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  createTakeout(request, shareID, options) {
    return FilesApiFp(this.configuration).createTakeout(request, shareID, options).then((request2) => request2(this.axios, this.basePath));
  }
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
    return FilesApiFp(this.configuration).deleteFiles(request, ignoreTrash, preserveFolder, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Download a file
   * @param {string} fileID File ID
   * @param {string} [shareID] Share ID
   * @param {string} [format] File format conversion
   * @param {boolean} [isTakeout] Is this a takeout file
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  downloadFile(fileID, shareID, format, isTakeout, options) {
    return FilesApiFp(this.configuration).downloadFile(fileID, shareID, format, isTakeout, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get information about a file
   * @param {string} fileID File ID
   * @param {string} [shareID] Share ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getFile(fileID, shareID, options) {
    return FilesApiFp(this.configuration).getFile(fileID, shareID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get the statistics of a file
   * @param {string} fileID File ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getFileStats(fileID, options) {
    return FilesApiFp(this.configuration).getFileStats(fileID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get the text of a text file
   * @param {string} fileID File ID
   * @param {string} [shareID] Share ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getFileText(fileID, shareID, options) {
    return FilesApiFp(this.configuration).getFileText(fileID, shareID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get files shared with the logged in user
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getSharedFiles(options) {
    return FilesApiFp(this.configuration).getSharedFiles(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get the result of an upload task. This will block until the upload is complete
   * @param {string} uploadID Upload ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getUploadResult(uploadID, options) {
    return FilesApiFp(this.configuration).getUploadResult(uploadID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Move a list of files to a new parent folder
   * @param {MoveFilesParams} request Move files request body
   * @param {string} [shareID] Share ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  moveFiles(request, shareID, options) {
    return FilesApiFp(this.configuration).moveFiles(request, shareID, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary structsore files from some time in the past
   * @param {RestoreFilesBody} request RestoreFiles files request body
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  restoreFiles(request, options) {
    return FilesApiFp(this.configuration).restoreFiles(request, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Search for files by filename
   * @param {string} search Filename to search for
   * @param {string} [baseFolderID] The folder to search in, defaults to the user\&#39;s home folder
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  searchByFilename(search, baseFolderID, options) {
    return FilesApiFp(this.configuration).searchByFilename(search, baseFolderID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Begin a new upload task
   * @param {NewUploadParams} request New upload request body
   * @param {string} [shareID] Share ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  startUpload(request, shareID, options) {
    return FilesApiFp(this.configuration).startUpload(request, shareID, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Move a list of files out of the trash, structsoring them to where they were before
   * @param {FilesListParams} request Un-trash files request body
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  unTrashFiles(request, options) {
    return FilesApiFp(this.configuration).unTrashFiles(request, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Update a File
   * @param {string} fileID File ID
   * @param {UpdateFileParams} request Update file request body
   * @param {string} [shareID] Share ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  updateFile(fileID, request, shareID, options) {
    return FilesApiFp(this.configuration).updateFile(fileID, request, shareID, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Add a chunk to a file upload
   * @param {string} uploadID Upload ID
   * @param {string} fileID File ID
   * @param {File} chunk File chunk
   * @param {string} [shareID] Share ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  uploadFileChunk(uploadID, fileID, chunk, shareID, options) {
    return FilesApiFp(this.configuration).uploadFileChunk(uploadID, fileID, chunk, shareID, options).then((request) => request(this.axios, this.basePath));
  }
};
var FolderApiAxiosParamCreator = function(configuration) {
  return {
    /**
     * 
     * @summary Create a new folder
     * @param {CreateFolderBody} request New folder body
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createFolder: async (request, shareID, options = {}) => {
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
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} folderID Folder ID
     * @param {string} [shareID] Share ID
     * @param {number} [timestamp] Past timestamp to view the folder at, in ms since epoch
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFolder: async (folderID, shareID, timestamp, options = {}) => {
      assertParamExists("getFolder", "folderID", folderID);
      const localVarPath = `/folder/{folderID}`.replace(`{${"folderID"}}`, encodeURIComponent(String(folderID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
      }
      if (timestamp !== void 0) {
        localVarQueryParameter["timestamp"] = timestamp;
      }
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} fileID File ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFolderHistory: async (fileID, options = {}) => {
      assertParamExists("getFolderHistory", "fileID", fileID);
      const localVarPath = `/files/{fileID}/history`.replace(`{${"fileID"}}`, encodeURIComponent(String(fileID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Accept"] = "*/*";
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
     * @param {string} folderID Folder ID
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    scanFolder: async (folderID, shareID, options = {}) => {
      assertParamExists("scanFolder", "folderID", folderID);
      const localVarPath = `/folder/{folderID}/scan`.replace(`{${"folderID"}}`, encodeURIComponent(String(folderID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
      }
      localVarHeaderParameter["Accept"] = "*/*";
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
     * @summary Set the cover image of a folder
     * @param {string} folderID Folder ID
     * @param {string} mediaID Media ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setFolderCover: async (folderID, mediaID, options = {}) => {
      assertParamExists("setFolderCover", "folderID", folderID);
      assertParamExists("setFolderCover", "mediaID", mediaID);
      const localVarPath = `/folder/{folderID}/cover`.replace(`{${"folderID"}}`, encodeURIComponent(String(folderID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (mediaID !== void 0) {
        localVarQueryParameter["mediaID"] = mediaID;
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
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async createFolder(request, shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.createFolder(request, shareID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FolderApi.createFolder"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get a folder
     * @param {string} folderID Folder ID
     * @param {string} [shareID] Share ID
     * @param {number} [timestamp] Past timestamp to view the folder at, in ms since epoch
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFolder(folderID, shareID, timestamp, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFolder(folderID, shareID, timestamp, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FolderApi.getFolder"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get actions of a folder at a given time
     * @param {string} fileID File ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFolderHistory(fileID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFolderHistory(fileID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FolderApi.getFolderHistory"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Dispatch a folder scan
     * @param {string} folderID Folder ID
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async scanFolder(folderID, shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.scanFolder(folderID, shareID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["FolderApi.scanFolder"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Set the cover image of a folder
     * @param {string} folderID Folder ID
     * @param {string} mediaID Media ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async setFolderCover(folderID, mediaID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.setFolderCover(folderID, mediaID, options);
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
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    createFolder(request, shareID, options) {
      return localVarFp.createFolder(request, shareID, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Get a folder
     * @param {string} folderID Folder ID
     * @param {string} [shareID] Share ID
     * @param {number} [timestamp] Past timestamp to view the folder at, in ms since epoch
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFolder(folderID, shareID, timestamp, options) {
      return localVarFp.getFolder(folderID, shareID, timestamp, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get actions of a folder at a given time
     * @param {string} fileID File ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFolderHistory(fileID, options) {
      return localVarFp.getFolderHistory(fileID, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Dispatch a folder scan
     * @param {string} folderID Folder ID
     * @param {string} [shareID] Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    scanFolder(folderID, shareID, options) {
      return localVarFp.scanFolder(folderID, shareID, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Set the cover image of a folder
     * @param {string} folderID Folder ID
     * @param {string} mediaID Media ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setFolderCover(folderID, mediaID, options) {
      return localVarFp.setFolderCover(folderID, mediaID, options).then((request) => request(axios, basePath));
    }
  };
};
var FolderApi = class extends BaseAPI {
  /**
   * 
   * @summary Create a new folder
   * @param {CreateFolderBody} request New folder body
   * @param {string} [shareID] Share ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  createFolder(request, shareID, options) {
    return FolderApiFp(this.configuration).createFolder(request, shareID, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get a folder
   * @param {string} folderID Folder ID
   * @param {string} [shareID] Share ID
   * @param {number} [timestamp] Past timestamp to view the folder at, in ms since epoch
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getFolder(folderID, shareID, timestamp, options) {
    return FolderApiFp(this.configuration).getFolder(folderID, shareID, timestamp, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get actions of a folder at a given time
   * @param {string} fileID File ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getFolderHistory(fileID, options) {
    return FolderApiFp(this.configuration).getFolderHistory(fileID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Dispatch a folder scan
   * @param {string} folderID Folder ID
   * @param {string} [shareID] Share ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  scanFolder(folderID, shareID, options) {
    return FolderApiFp(this.configuration).scanFolder(folderID, shareID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Set the cover image of a folder
   * @param {string} folderID Folder ID
   * @param {string} mediaID Media ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  setFolderCover(folderID, mediaID, options) {
    return FolderApiFp(this.configuration).setFolderCover(folderID, mediaID, options).then((request) => request(this.axios, this.basePath));
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
     * @param {string} [username] Username of owner whose media to drop. If empty, drops all media.
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    dropMedia: async (username, options = {}) => {
      const localVarPath = `/media/drop`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (username !== void 0) {
        localVarQueryParameter["username"] = username;
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
     * @summary Get paginated media
     * @param {MediaBatchParams} request Media Batch Params
     * @param {string} [shareID] File ShareID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMedia: async (request, shareID, options = {}) => {
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
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
      }
      localVarHeaderParameter["Content-Type"] = "application/json";
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} mediaID ID of media
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaFile: async (mediaID, options = {}) => {
      assertParamExists("getMediaFile", "mediaID", mediaID);
      const localVarPath = `/media/{mediaID}/file`.replace(`{${"mediaID"}}`, encodeURIComponent(String(mediaID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} mediaID Media ID
     * @param {string} extension Extension
     * @param {GetMediaImageQualityEnum} quality Image Quality
     * @param {number} [page] Page number
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaImage: async (mediaID, extension, quality, page, options = {}) => {
      assertParamExists("getMediaImage", "mediaID", mediaID);
      assertParamExists("getMediaImage", "extension", extension);
      assertParamExists("getMediaImage", "quality", quality);
      const localVarPath = `/media/{mediaID}.{extension}`.replace(`{${"mediaID"}}`, encodeURIComponent(String(mediaID))).replace(`{${"extension"}}`, encodeURIComponent(String(extension)));
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
      localVarHeaderParameter["Accept"] = "image/*";
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
     * @param {string} mediaID Media ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaInfo: async (mediaID, options = {}) => {
      assertParamExists("getMediaInfo", "mediaID", mediaID);
      const localVarPath = `/media/{mediaID}/info`.replace(`{${"mediaID"}}`, encodeURIComponent(String(mediaID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Accept"] = "application/json";
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
      localVarHeaderParameter["Accept"] = "application/json";
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
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} mediaID ID of media
     * @param {boolean} liked Liked status to set
     * @param {string} [shareID] ShareID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setMediaLiked: async (mediaID, liked, shareID, options = {}) => {
      assertParamExists("setMediaLiked", "mediaID", mediaID);
      assertParamExists("setMediaLiked", "liked", liked);
      const localVarPath = `/media/{mediaID}/liked`.replace(`{${"mediaID"}}`, encodeURIComponent(String(mediaID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (shareID !== void 0) {
        localVarQueryParameter["shareID"] = shareID;
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
     * @param {MediaIDsParams} mediaIDs MediaIDs to change visibility of
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setMediaVisibility: async (hidden, mediaIDs, options = {}) => {
      assertParamExists("setMediaVisibility", "hidden", hidden);
      assertParamExists("setMediaVisibility", "mediaIDs", mediaIDs);
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
      localVarRequestOptions.data = serializeDataIfNeeded(mediaIDs, localVarRequestOptions, configuration);
      return {
        url: toPathString(localVarUrlObj),
        options: localVarRequestOptions
      };
    },
    /**
     * 
     * @summary Stream a video
     * @param {string} mediaID ID of media
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    streamVideo: async (mediaID, options = {}) => {
      assertParamExists("streamVideo", "mediaID", mediaID);
      const localVarPath = `/media/{mediaID}/video`.replace(`{${"mediaID"}}`, encodeURIComponent(String(mediaID)));
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
     * @param {string} [username] Username of owner whose media to drop. If empty, drops all media.
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async dropMedia(username, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.dropMedia(username, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.dropMedia"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get paginated media
     * @param {MediaBatchParams} request Media Batch Params
     * @param {string} [shareID] File ShareID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getMedia(request, shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getMedia(request, shareID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.getMedia"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get file of media by id
     * @param {string} mediaID ID of media
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getMediaFile(mediaID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getMediaFile(mediaID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.getMediaFile"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get a media image bytes
     * @param {string} mediaID Media ID
     * @param {string} extension Extension
     * @param {GetMediaImageQualityEnum} quality Image Quality
     * @param {number} [page] Page number
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getMediaImage(mediaID, extension, quality, page, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getMediaImage(mediaID, extension, quality, page, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.getMediaImage"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get media info
     * @param {string} mediaID Media ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getMediaInfo(mediaID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getMediaInfo(mediaID, options);
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
     * @param {string} mediaID ID of media
     * @param {boolean} liked Liked status to set
     * @param {string} [shareID] ShareID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async setMediaLiked(mediaID, liked, shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.setMediaLiked(mediaID, liked, shareID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.setMediaLiked"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Set media visibility
     * @param {boolean} hidden Set the media visibility
     * @param {MediaIDsParams} mediaIDs MediaIDs to change visibility of
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async setMediaVisibility(hidden, mediaIDs, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.setMediaVisibility(hidden, mediaIDs, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["MediaApi.setMediaVisibility"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Stream a video
     * @param {string} mediaID ID of media
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async streamVideo(mediaID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.streamVideo(mediaID, options);
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
     * @param {string} [username] Username of owner whose media to drop. If empty, drops all media.
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    dropMedia(username, options) {
      return localVarFp.dropMedia(username, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get paginated media
     * @param {MediaBatchParams} request Media Batch Params
     * @param {string} [shareID] File ShareID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMedia(request, shareID, options) {
      return localVarFp.getMedia(request, shareID, options).then((request2) => request2(axios, basePath));
    },
    /**
     * 
     * @summary Get file of media by id
     * @param {string} mediaID ID of media
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaFile(mediaID, options) {
      return localVarFp.getMediaFile(mediaID, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get a media image bytes
     * @param {string} mediaID Media ID
     * @param {string} extension Extension
     * @param {GetMediaImageQualityEnum} quality Image Quality
     * @param {number} [page] Page number
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaImage(mediaID, extension, quality, page, options) {
      return localVarFp.getMediaImage(mediaID, extension, quality, page, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get media info
     * @param {string} mediaID Media ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getMediaInfo(mediaID, options) {
      return localVarFp.getMediaInfo(mediaID, options).then((request) => request(axios, basePath));
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
     * @param {string} mediaID ID of media
     * @param {boolean} liked Liked status to set
     * @param {string} [shareID] ShareID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setMediaLiked(mediaID, liked, shareID, options) {
      return localVarFp.setMediaLiked(mediaID, liked, shareID, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Set media visibility
     * @param {boolean} hidden Set the media visibility
     * @param {MediaIDsParams} mediaIDs MediaIDs to change visibility of
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setMediaVisibility(hidden, mediaIDs, options) {
      return localVarFp.setMediaVisibility(hidden, mediaIDs, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Stream a video
     * @param {string} mediaID ID of media
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    streamVideo(mediaID, options) {
      return localVarFp.streamVideo(mediaID, options).then((request) => request(axios, basePath));
    }
  };
};
var MediaApi = class extends BaseAPI {
  /**
   * 
   * @summary Make sure all media is correctly synced with the file system
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  cleanupMedia(options) {
    return MediaApiFp(this.configuration).cleanupMedia(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Drop all computed media HDIR data. Must be server owner.
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  dropHDIRs(options) {
    return MediaApiFp(this.configuration).dropHDIRs(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary DANGEROUS. Drop all computed media and clear thumbnail in-memory and filesystem cache. Must be server owner.
   * @param {string} [username] Username of owner whose media to drop. If empty, drops all media.
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  dropMedia(username, options) {
    return MediaApiFp(this.configuration).dropMedia(username, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get paginated media
   * @param {MediaBatchParams} request Media Batch Params
   * @param {string} [shareID] File ShareID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getMedia(request, shareID, options) {
    return MediaApiFp(this.configuration).getMedia(request, shareID, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get file of media by id
   * @param {string} mediaID ID of media
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getMediaFile(mediaID, options) {
    return MediaApiFp(this.configuration).getMediaFile(mediaID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get a media image bytes
   * @param {string} mediaID Media ID
   * @param {string} extension Extension
   * @param {GetMediaImageQualityEnum} quality Image Quality
   * @param {number} [page] Page number
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getMediaImage(mediaID, extension, quality, page, options) {
    return MediaApiFp(this.configuration).getMediaImage(mediaID, extension, quality, page, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get media info
   * @param {string} mediaID Media ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getMediaInfo(mediaID, options) {
    return MediaApiFp(this.configuration).getMediaInfo(mediaID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get media type dictionary
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
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
   */
  getRandomMedia(count, options) {
    return MediaApiFp(this.configuration).getRandomMedia(count, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Like a media
   * @param {string} mediaID ID of media
   * @param {boolean} liked Liked status to set
   * @param {string} [shareID] ShareID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  setMediaLiked(mediaID, liked, shareID, options) {
    return MediaApiFp(this.configuration).setMediaLiked(mediaID, liked, shareID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Set media visibility
   * @param {boolean} hidden Set the media visibility
   * @param {MediaIDsParams} mediaIDs MediaIDs to change visibility of
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  setMediaVisibility(hidden, mediaIDs, options) {
    return MediaApiFp(this.configuration).setMediaVisibility(hidden, mediaIDs, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Stream a video
   * @param {string} mediaID ID of media
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  streamVideo(mediaID, options) {
    return MediaApiFp(this.configuration).streamVideo(mediaID, options).then((request) => request(this.axios, this.basePath));
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
     * @param {string} shareID Share ID
     * @param {AddUserParams} request Share Accessors
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    addUserToShare: async (shareID, request, options = {}) => {
      assertParamExists("addUserToShare", "shareID", shareID);
      assertParamExists("addUserToShare", "request", request);
      const localVarPath = `/share/{shareID}/accessors`.replace(`{${"shareID"}}`, encodeURIComponent(String(shareID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "POST" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Content-Type"] = "application/json";
      localVarHeaderParameter["Accept"] = "application/json";
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
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} shareID Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteFileShare: async (shareID, options = {}) => {
      assertParamExists("deleteFileShare", "shareID", shareID);
      const localVarPath = `/share/{shareID}`.replace(`{${"shareID"}}`, encodeURIComponent(String(shareID)));
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
     * @param {string} shareID Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFileShare: async (shareID, options = {}) => {
      assertParamExists("getFileShare", "shareID", shareID);
      const localVarPath = `/share/{shareID}`.replace(`{${"shareID"}}`, encodeURIComponent(String(shareID)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} shareID Share ID
     * @param {string} username Username
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    removeUserFromShare: async (shareID, username, options = {}) => {
      assertParamExists("removeUserFromShare", "shareID", shareID);
      assertParamExists("removeUserFromShare", "username", username);
      const localVarPath = `/share/{shareID}/accessors/{username}`.replace(`{${"shareID"}}`, encodeURIComponent(String(shareID))).replace(`{${"username"}}`, encodeURIComponent(String(username)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "DELETE" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} shareID Share ID
     * @param {boolean} _public Share Public Status
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setSharePublic: async (shareID, _public, options = {}) => {
      assertParamExists("setSharePublic", "shareID", shareID);
      assertParamExists("setSharePublic", "_public", _public);
      const localVarPath = `/share/{shareID}/public`.replace(`{${"shareID"}}`, encodeURIComponent(String(shareID)));
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
     * @param {string} shareID Share ID
     * @param {string} username Username
     * @param {PermissionsParams} request Share Permissions Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    updateShareAccessorPermissions: async (shareID, username, request, options = {}) => {
      assertParamExists("updateShareAccessorPermissions", "shareID", shareID);
      assertParamExists("updateShareAccessorPermissions", "username", username);
      assertParamExists("updateShareAccessorPermissions", "request", request);
      const localVarPath = `/share/{shareID}/accessors/{username}`.replace(`{${"shareID"}}`, encodeURIComponent(String(shareID))).replace(`{${"username"}}`, encodeURIComponent(String(username)));
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "PATCH" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Content-Type"] = "application/json";
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} shareID Share ID
     * @param {AddUserParams} request Share Accessors
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async addUserToShare(shareID, request, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.addUserToShare(shareID, request, options);
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
     * @param {string} shareID Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async deleteFileShare(shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.deleteFileShare(shareID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ShareApi.deleteFileShare"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Get a file share
     * @param {string} shareID Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getFileShare(shareID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getFileShare(shareID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ShareApi.getFileShare"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Remove a user from a file share
     * @param {string} shareID Share ID
     * @param {string} username Username
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async removeUserFromShare(shareID, username, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.removeUserFromShare(shareID, username, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ShareApi.removeUserFromShare"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Update a share\'s \"public\" status
     * @param {string} shareID Share ID
     * @param {boolean} _public Share Public Status
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async setSharePublic(shareID, _public, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.setSharePublic(shareID, _public, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["ShareApi.setSharePublic"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Update a share\'s user permissions
     * @param {string} shareID Share ID
     * @param {string} username Username
     * @param {PermissionsParams} request Share Permissions Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async updateShareAccessorPermissions(shareID, username, request, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.updateShareAccessorPermissions(shareID, username, request, options);
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
     * @param {string} shareID Share ID
     * @param {AddUserParams} request Share Accessors
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    addUserToShare(shareID, request, options) {
      return localVarFp.addUserToShare(shareID, request, options).then((request2) => request2(axios, basePath));
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
     * @param {string} shareID Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteFileShare(shareID, options) {
      return localVarFp.deleteFileShare(shareID, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Get a file share
     * @param {string} shareID Share ID
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getFileShare(shareID, options) {
      return localVarFp.getFileShare(shareID, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Remove a user from a file share
     * @param {string} shareID Share ID
     * @param {string} username Username
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    removeUserFromShare(shareID, username, options) {
      return localVarFp.removeUserFromShare(shareID, username, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Update a share\'s \"public\" status
     * @param {string} shareID Share ID
     * @param {boolean} _public Share Public Status
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    setSharePublic(shareID, _public, options) {
      return localVarFp.setSharePublic(shareID, _public, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Update a share\'s user permissions
     * @param {string} shareID Share ID
     * @param {string} username Username
     * @param {PermissionsParams} request Share Permissions Params
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    updateShareAccessorPermissions(shareID, username, request, options) {
      return localVarFp.updateShareAccessorPermissions(shareID, username, request, options).then((request2) => request2(axios, basePath));
    }
  };
};
var ShareApi = class extends BaseAPI {
  /**
   * 
   * @summary Add a user to a file share
   * @param {string} shareID Share ID
   * @param {AddUserParams} request Share Accessors
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  addUserToShare(shareID, request, options) {
    return ShareApiFp(this.configuration).addUserToShare(shareID, request, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Share a file
   * @param {FileShareParams} request New File Share Params
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  createFileShare(request, options) {
    return ShareApiFp(this.configuration).createFileShare(request, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Delete a file share
   * @param {string} shareID Share ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  deleteFileShare(shareID, options) {
    return ShareApiFp(this.configuration).deleteFileShare(shareID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get a file share
   * @param {string} shareID Share ID
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getFileShare(shareID, options) {
    return ShareApiFp(this.configuration).getFileShare(shareID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Remove a user from a file share
   * @param {string} shareID Share ID
   * @param {string} username Username
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  removeUserFromShare(shareID, username, options) {
    return ShareApiFp(this.configuration).removeUserFromShare(shareID, username, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Update a share\'s \"public\" status
   * @param {string} shareID Share ID
   * @param {boolean} _public Share Public Status
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  setSharePublic(shareID, _public, options) {
    return ShareApiFp(this.configuration).setSharePublic(shareID, _public, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Update a share\'s user permissions
   * @param {string} shareID Share ID
   * @param {string} username Username
   * @param {PermissionsParams} request Share Permissions Params
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  updateShareAccessorPermissions(shareID, username, request, options) {
    return ShareApiFp(this.configuration).updateShareAccessorPermissions(shareID, username, request, options).then((request2) => request2(this.axios, this.basePath));
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
      localVarHeaderParameter["Content-Type"] = "application/json";
      localVarHeaderParameter["Accept"] = "*/*";
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
     * @param {string} serverID Server ID to delete
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteRemote: async (serverID, options = {}) => {
      assertParamExists("deleteRemote", "serverID", serverID);
      const localVarPath = `/tower/{serverID}`.replace(`{${"serverID"}}`, encodeURIComponent(String(serverID)));
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
     * @summary Enable trace logging
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    enableTraceLogging: async (options = {}) => {
      const localVarPath = `/tower/trace`;
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
     * @summary Flush Cache
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    flushCache: async (options = {}) => {
      const localVarPath = `/tower/cache`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "DELETE" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Accept"] = "application/json";
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
      if (timestamp !== void 0) {
        localVarQueryParameter["timestamp"] = timestamp;
      }
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @summary Get a page of file actions
     * @param {number} [page] Page number
     * @param {number} [pageSize] Number of items per page
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getPagedHistoryActions: async (page, pageSize, options = {}) => {
      const localVarPath = `/tower/history`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      if (page !== void 0) {
        localVarQueryParameter["page"] = page;
      }
      if (pageSize !== void 0) {
        localVarQueryParameter["pageSize"] = pageSize;
      }
      localVarHeaderParameter["Accept"] = "application/json";
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
      localVarHeaderParameter["Accept"] = "*/*";
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
     * @summary Get Running Tasks
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getRunningTasks: async (options = {}) => {
      const localVarPath = `/tower/tasks`;
      const localVarUrlObj = new URL(localVarPath, DUMMY_BASE_URL);
      let baseOptions;
      if (configuration) {
        baseOptions = configuration.baseOptions;
      }
      const localVarRequestOptions = __spreadValues(__spreadValues({ method: "GET" }, baseOptions), options);
      const localVarHeaderParameter = {};
      const localVarQueryParameter = {};
      localVarHeaderParameter["Accept"] = "application/json";
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
      localVarHeaderParameter["Accept"] = "application/json";
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
      localVarHeaderParameter["Accept"] = "application/json";
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
     * @param {string} serverID Server ID of the tower to back up
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    launchBackup: async (serverID, options = {}) => {
      assertParamExists("launchBackup", "serverID", serverID);
      const localVarPath = `/tower/{serverID}/backup`.replace(`{${"serverID"}}`, encodeURIComponent(String(serverID)));
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
     * @param {string} serverID Server ID to delete
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async deleteRemote(serverID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.deleteRemote(serverID, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.deleteRemote"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Enable trace logging
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async enableTraceLogging(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.enableTraceLogging(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.enableTraceLogging"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
      return (axios, basePath) => createRequestFunction(localVarAxiosArgs, globalAxios2, BASE_PATH, configuration)(axios, localVarOperationServerBasePath || basePath);
    },
    /**
     * 
     * @summary Flush Cache
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async flushCache(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.flushCache(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.flushCache"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
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
     * @summary Get a page of file actions
     * @param {number} [page] Page number
     * @param {number} [pageSize] Number of items per page
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getPagedHistoryActions(page, pageSize, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getPagedHistoryActions(page, pageSize, options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.getPagedHistoryActions"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
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
     * @summary Get Running Tasks
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async getRunningTasks(options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.getRunningTasks(options);
      const localVarOperationServerIndex = (_a = configuration == null ? void 0 : configuration.serverIndex) != null ? _a : 0;
      const localVarOperationServerBasePath = (_c = (_b = operationServerMap["TowersApi.getRunningTasks"]) == null ? void 0 : _b[localVarOperationServerIndex]) == null ? void 0 : _c.url;
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
     * @param {string} serverID Server ID of the tower to back up
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    async launchBackup(serverID, options) {
      var _a, _b, _c;
      const localVarAxiosArgs = await localVarAxiosParamCreator.launchBackup(serverID, options);
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
     * @param {string} serverID Server ID to delete
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    deleteRemote(serverID, options) {
      return localVarFp.deleteRemote(serverID, options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Enable trace logging
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    enableTraceLogging(options) {
      return localVarFp.enableTraceLogging(options).then((request) => request(axios, basePath));
    },
    /**
     * 
     * @summary Flush Cache
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    flushCache(options) {
      return localVarFp.flushCache(options).then((request) => request(axios, basePath));
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
     * @summary Get a page of file actions
     * @param {number} [page] Page number
     * @param {number} [pageSize] Number of items per page
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getPagedHistoryActions(page, pageSize, options) {
      return localVarFp.getPagedHistoryActions(page, pageSize, options).then((request) => request(axios, basePath));
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
     * @summary Get Running Tasks
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    getRunningTasks(options) {
      return localVarFp.getRunningTasks(options).then((request) => request(axios, basePath));
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
     * @param {string} serverID Server ID of the tower to back up
     * @param {*} [options] Override http request option.
     * @throws {RequiredError}
     */
    launchBackup(serverID, options) {
      return localVarFp.launchBackup(serverID, options).then((request) => request(axios, basePath));
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
   */
  createRemote(request, options) {
    return TowersApiFp(this.configuration).createRemote(request, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Delete a remote
   * @param {string} serverID Server ID to delete
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  deleteRemote(serverID, options) {
    return TowersApiFp(this.configuration).deleteRemote(serverID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Enable trace logging
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  enableTraceLogging(options) {
    return TowersApiFp(this.configuration).enableTraceLogging(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Flush Cache
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  flushCache(options) {
    return TowersApiFp(this.configuration).flushCache(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get information about a file
   * @param {string} timestamp Timestamp in milliseconds since epoch
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getBackupInfo(timestamp, options) {
    return TowersApiFp(this.configuration).getBackupInfo(timestamp, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get a page of file actions
   * @param {number} [page] Page number
   * @param {number} [pageSize] Number of items per page
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getPagedHistoryActions(page, pageSize, options) {
    return TowersApiFp(this.configuration).getPagedHistoryActions(page, pageSize, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get all remotes
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getRemotes(options) {
    return TowersApiFp(this.configuration).getRemotes(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get Running Tasks
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getRunningTasks(options) {
    return TowersApiFp(this.configuration).getRunningTasks(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get server info
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
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
   */
  initializeTower(request, options) {
    return TowersApiFp(this.configuration).initializeTower(request, options).then((request2) => request2(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Launch backup on a tower
   * @param {string} serverID Server ID of the tower to back up
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  launchBackup(serverID, options) {
    return TowersApiFp(this.configuration).launchBackup(serverID, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Reset tower
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
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
      if (setActive !== void 0) {
        localVarQueryParameter["setActive"] = setActive;
      }
      localVarHeaderParameter["Accept"] = "application/json";
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
      if (newFullName !== void 0) {
        localVarQueryParameter["newFullName"] = newFullName;
      }
      localVarHeaderParameter["Accept"] = "application/json";
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
      localVarHeaderParameter["Accept"] = "application/json";
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
      localVarHeaderParameter["Accept"] = "application/json";
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
      localVarHeaderParameter["Accept"] = "application/json";
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
      localVarHeaderParameter["Accept"] = "application/json";
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
      if (setAdmin !== void 0) {
        localVarQueryParameter["setAdmin"] = setAdmin;
      }
      localVarHeaderParameter["Accept"] = "application/json";
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
      localVarHeaderParameter["Content-Type"] = "application/json";
      localVarHeaderParameter["Accept"] = "application/json";
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
   */
  deleteUser(username, options) {
    return UsersApiFp(this.configuration).deleteUser(username, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Gets the user based on the auth token
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
   */
  getUser(options) {
    return UsersApiFp(this.configuration).getUser(options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Get all users, including (possibly) sensitive information like password hashes
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
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
   */
  loginUser(loginParams, options) {
    return UsersApiFp(this.configuration).loginUser(loginParams, options).then((request) => request(this.axios, this.basePath));
  }
  /**
   * 
   * @summary Logout User
   * @param {*} [options] Override http request option.
   * @throws {RequiredError}
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
   */
  updateUserPassword(username, passwordUpdateParams, options) {
    return UsersApiFp(this.configuration).updateUserPassword(username, passwordUpdateParams, options).then((request) => request(this.axios, this.basePath));
  }
};

// AllApi.ts
function WeblensAPIFactory(apiEndpoint) {
  return {
    MediaAPI: MediaApiFactory({}, apiEndpoint),
    FilesAPI: FilesApiFactory({}, apiEndpoint),
    FoldersAPI: FolderApiFactory({}, apiEndpoint),
    TowersAPI: TowersApiFactory({}, apiEndpoint),
    SharesAPI: ShareApiFactory({}, apiEndpoint),
    UsersAPI: UsersApiFactory({}, apiEndpoint),
    APIKeysAPI: APIKeysApiFactory({}, apiEndpoint),
    FeatureFlagsAPI: FeatureFlagsApiFactory({}, apiEndpoint)
  };
}
export {
  APIKeysApi,
  APIKeysApiAxiosParamCreator,
  APIKeysApiFactory,
  APIKeysApiFp,
  FeatureFlagsApi,
  FeatureFlagsApiAxiosParamCreator,
  FeatureFlagsApiFactory,
  FeatureFlagsApiFp,
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
  WeblensAPIFactory
};
