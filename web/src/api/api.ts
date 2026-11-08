/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export interface CreateSandboxReq {
  editable?: boolean;
  name: string;
  placeholder?: string;
  templateID: number;
}

export interface CreateTemplateReq {
  /** @default false */
  internetAccess?: boolean;
  /** Supported languages: c, go, php, python, javascript. */
  language: string[];
  /**
   * @min 0
   * @max 4096
   * @default 0
   */
  maxMemory?: number;
  name: string;
  /**
   * @min 0
   * @max 60
   * @default 0
   */
  timeout?: number;
}

export interface ErrorResponse {
  msg: string;
}

export interface ExecutionResult {
  end_at: number;
  result: RuntimeExecOutput;
  start_at: number;
}

export interface ListSandboxesResp {
  sandboxes: Sandbox[];
  total: number;
}

export interface ListTemplatesResp {
  templates: Template[];
  total: number;
}

export interface Sandbox {
  createdAt: string;
  editable: boolean;
  id: number;
  name: string;
  placeholder: string;
  template?: Template | null;
  templateID: number;
  uid: string;
  updatedAt: string;
}

export interface SignInForm {
  /** Password must contain 8-72 UTF-8 bytes. */
  password: string;
}

export interface Template {
  createdAt: string;
  id: number;
  internetAccess: boolean;
  language?: string[] | null;
  /** Keep the legacy CPU quota visible until the template is migrated. */
  maxCpus: number;
  maxMemory: number;
  name: string;
  timeout: number;
  updatedAt: string;
}

export interface UpdateSandboxReq {
  editable?: boolean;
  name: string;
  placeholder?: string;
  templateID: number;
}

export interface UpdateTemplateReq {
  /** @default false */
  internetAccess?: boolean;
  /** Supported languages: c, go, php, python, javascript. */
  language: string[];
  /**
   * @min 0
   * @max 4096
   * @default 0
   */
  maxMemory?: number;
  name: string;
  /**
   * @min 0
   * @max 60
   * @default 0
   */
  timeout?: number;
}

export interface DtoResponseExecutionResult {
  data: ExecutionResult;
}

export interface DtoResponseListSandboxesResp {
  data: ListSandboxesResp;
}

export interface DtoResponseListTemplatesResp {
  data: ListTemplatesResp;
}

export interface DtoResponseSandbox {
  data: Sandbox;
}

export interface DtoResponseTemplate {
  data: Template;
}

export interface DtoResponseAny {
  data: any;
}

export interface DtoResponseArrayTemplate {
  data: Template[];
}

export interface DtoResponseString {
  data: string;
}

export interface RuntimeExecOutput {
  duration: number;
  error: string;
  stderr: string;
  stdout: string;
}

export type ProfileData = DtoResponseAny;

export type SignInData = DtoResponseString;

export type SignOutData = DtoResponseString;

export interface GetSandboxParams {
  /** ID */
  id: number;
}

export type GetSandboxData = DtoResponseSandbox;

export interface UpdateSandboxParams {
  /** ID */
  id: number;
}

export type UpdateSandboxData = DtoResponseString;

export interface DeleteSandboxParams {
  /** ID */
  id: number;
}

export type DeleteSandboxData = DtoResponseString;

export interface ListSandboxesParams {
  /**
   * Page number; values below 1 use 1
   * @default 1
   */
  page?: number;
  /**
   * Page size; values below 1 use 20
   * @default 20
   */
  pageSize?: number;
}

export type ListSandboxesData = DtoResponseListSandboxesResp;

export type CreateSandboxData = DtoResponseSandbox;

export interface GetTemplateParams {
  /** ID */
  id: number;
}

export type GetTemplateData = DtoResponseTemplate;

export interface UpdateTemplateParams {
  /** ID */
  id: number;
}

export type UpdateTemplateData = DtoResponseString;

export interface DeleteTemplateParams {
  /** ID */
  id: number;
}

export type DeleteTemplateData = DtoResponseString;

export interface ListTemplatesParams {
  /**
   * Page number; values below 1 use 1
   * @default 1
   */
  page?: number;
  /**
   * Page size; values below 1 use 20
   * @default 20
   */
  pageSize?: number;
}

export type ListTemplatesData = DtoResponseListTemplatesResp;

export type CreateTemplateData = DtoResponseTemplate;

export type AllTemplatesData = DtoResponseArrayTemplate;

export interface ExecuteCodePayload {
  /** Programming language */
  lang?: string;
  /** Source code */
  code?: string;
}

export interface ExecuteCodeParams {
  /** UID */
  uid: string;
}

export type ExecuteCodeData = DtoResponseExecutionResult;

import type {
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  HeadersDefaults,
  ResponseType,
} from "axios";
import axios from "axios";

export type QueryParamsType = Record<string | number, any>;

export interface FullRequestParams
  extends Omit<AxiosRequestConfig, "data" | "params" | "url" | "responseType"> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseType;
  /** request body */
  body?: unknown;
}

export type RequestParams = Omit<
  FullRequestParams,
  "body" | "method" | "query" | "path"
>;

export interface ApiConfig<SecurityDataType = unknown>
  extends Omit<AxiosRequestConfig, "data" | "cancelToken"> {
  securityWorker?: (
    securityData: SecurityDataType | null,
  ) => Promise<AxiosRequestConfig | void> | AxiosRequestConfig | void;
  secure?: boolean;
  format?: ResponseType;
}

export enum ContentType {
  Json = "application/json",
  JsonApi = "application/vnd.api+json",
  FormData = "multipart/form-data",
  UrlEncoded = "application/x-www-form-urlencoded",
  Text = "text/plain",
}

export class HttpClient<SecurityDataType = unknown> {
  public instance: AxiosInstance;
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>["securityWorker"];
  private secure?: boolean;
  private format?: ResponseType;

  constructor({
    securityWorker,
    secure,
    format,
    ...axiosConfig
  }: ApiConfig<SecurityDataType> = {}) {
    this.instance = axios.create({
      ...axiosConfig,
      baseURL: axiosConfig.baseURL || "/",
    });
    this.secure = secure;
    this.format = format;
    this.securityWorker = securityWorker;
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected mergeRequestParams(
    params1: AxiosRequestConfig,
    params2?: AxiosRequestConfig,
  ): AxiosRequestConfig {
    const method = params1.method || (params2 && params2.method);

    return {
      ...this.instance.defaults,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...((method &&
          this.instance.defaults.headers[
            method.toLowerCase() as keyof HeadersDefaults
          ]) ||
          {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    };
  }

  protected stringifyFormItem(formItem: unknown) {
    if (typeof formItem === "object" && formItem !== null) {
      return JSON.stringify(formItem);
    } else {
      return `${formItem}`;
    }
  }

  protected createFormData(input: Record<string, unknown>): FormData {
    if (input instanceof FormData) {
      return input;
    }
    return Object.keys(input || {}).reduce((formData, key) => {
      const property = input[key];
      const propertyContent: any[] =
        property instanceof Array ? property : [property];

      for (const formItem of propertyContent) {
        const isFileType = formItem instanceof Blob || formItem instanceof File;
        formData.append(
          key,
          isFileType ? formItem : this.stringifyFormItem(formItem),
        );
      }

      return formData;
    }, new FormData());
  }

  public request = async <T = any, _E = any>({
    secure,
    path,
    type,
    query,
    format,
    body,
    ...params
  }: FullRequestParams): Promise<AxiosResponse<T>> => {
    const secureParams =
      ((typeof secure === "boolean" ? secure : this.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const responseFormat = format || this.format || undefined;

    if (
      type === ContentType.FormData &&
      body &&
      body !== null &&
      typeof body === "object"
    ) {
      body = this.createFormData(body as Record<string, unknown>);
    }

    if (
      type === ContentType.Text &&
      body &&
      body !== null &&
      typeof body !== "string"
    ) {
      body = JSON.stringify(body);
    }

    return this.instance.request({
      ...requestParams,
      headers: {
        ...(requestParams.headers || {}),
        ...(type ? { "Content-Type": type } : {}),
      },
      params: query,
      responseType: responseFormat,
      data: body,
      url: path,
    });
  };
}

/**
 * @title Elaina API
 * @version 1.0
 * @baseUrl /
 * @contact
 *
 * Management and WASM code execution API. Successful JSON responses use a data envelope; errors contain msg.
 */
export class Api<
  SecurityDataType extends unknown,
> extends HttpClient<SecurityDataType> {
  auth = {
    /**
     * @description Returns data: null when the session is authenticated.
     *
     * @tags Auth
     * @name Profile
     * @summary Check the current session
     * @request GET:/api/auth/profile
     * @secure
     */
    profile: (params: RequestParams = {}) =>
      this.request<ProfileData, ErrorResponse>({
        path: `/api/auth/profile`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Auth
     * @name SignIn
     * @summary Sign in with the administrator password
     * @request POST:/api/auth/sign-in
     */
    signIn: (data: SignInForm, params: RequestParams = {}) =>
      this.request<SignInData, ErrorResponse>({
        path: `/api/auth/sign-in`,
        method: "POST",
        body: data,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Auth
     * @name SignOut
     * @summary Sign out the current session
     * @request POST:/api/auth/sign-out
     * @secure
     */
    signOut: (params: RequestParams = {}) =>
      this.request<SignOutData, ErrorResponse>({
        path: `/api/auth/sign-out`,
        method: "POST",
        secure: true,
        format: "json",
        ...params,
      }),
  };
  sandboxes = {
    /**
     * No description
     *
     * @tags Sandboxes
     * @name GetSandbox
     * @summary Get a sandbox
     * @request GET:/api/sandbox/{id}
     * @secure
     */
    getSandbox: ({ id }: GetSandboxParams, params: RequestParams = {}) =>
      this.request<GetSandboxData, ErrorResponse>({
        path: `/api/sandbox/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Sandboxes
     * @name UpdateSandbox
     * @summary Update a sandbox
     * @request PUT:/api/sandbox/{id}
     * @secure
     */
    updateSandbox: (
      { id }: UpdateSandboxParams,
      data: UpdateSandboxReq,
      params: RequestParams = {},
    ) =>
      this.request<UpdateSandboxData, ErrorResponse>({
        path: `/api/sandbox/${id}`,
        method: "PUT",
        body: data,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Sandboxes
     * @name DeleteSandbox
     * @summary Delete a sandbox
     * @request DELETE:/api/sandbox/{id}
     * @secure
     */
    deleteSandbox: ({ id }: DeleteSandboxParams, params: RequestParams = {}) =>
      this.request<DeleteSandboxData, ErrorResponse>({
        path: `/api/sandbox/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Sandboxes
     * @name ListSandboxes
     * @summary List sandboxes
     * @request GET:/api/sandboxes
     * @secure
     */
    listSandboxes: (
      query: ListSandboxesParams = {},
      params: RequestParams = {},
    ) =>
      this.request<ListSandboxesData, ErrorResponse>({
        path: `/api/sandboxes`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Sandboxes
     * @name CreateSandbox
     * @summary Create a sandbox
     * @request POST:/api/sandboxes
     * @secure
     */
    createSandbox: (data: CreateSandboxReq, params: RequestParams = {}) =>
      this.request<CreateSandboxData, ErrorResponse>({
        path: `/api/sandboxes`,
        method: "POST",
        body: data,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  templates = {
    /**
     * No description
     *
     * @tags Templates
     * @name GetTemplate
     * @summary Get a template
     * @request GET:/api/template/{id}
     * @secure
     */
    getTemplate: ({ id }: GetTemplateParams, params: RequestParams = {}) =>
      this.request<GetTemplateData, ErrorResponse>({
        path: `/api/template/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Zero timeout uses 10 seconds; zero memory uses 4096 MiB of WASM linear memory. Legacy CPU/container/DNS fields are accepted only when zero or empty.
     *
     * @tags Templates
     * @name UpdateTemplate
     * @summary Update a template
     * @request PUT:/api/template/{id}
     * @secure
     */
    updateTemplate: (
      { id }: UpdateTemplateParams,
      data: UpdateTemplateReq,
      params: RequestParams = {},
    ) =>
      this.request<UpdateTemplateData, ErrorResponse>({
        path: `/api/template/${id}`,
        method: "PUT",
        body: data,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Templates
     * @name DeleteTemplate
     * @summary Delete a template
     * @request DELETE:/api/template/{id}
     * @secure
     */
    deleteTemplate: (
      { id }: DeleteTemplateParams,
      params: RequestParams = {},
    ) =>
      this.request<DeleteTemplateData, ErrorResponse>({
        path: `/api/template/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Templates
     * @name ListTemplates
     * @summary List templates
     * @request GET:/api/templates
     * @secure
     */
    listTemplates: (
      query: ListTemplatesParams = {},
      params: RequestParams = {},
    ) =>
      this.request<ListTemplatesData, ErrorResponse>({
        path: `/api/templates`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Zero timeout uses 10 seconds; zero memory uses 4096 MiB of WASM linear memory. Legacy CPU/container/DNS fields are accepted only when zero or empty.
     *
     * @tags Templates
     * @name CreateTemplate
     * @summary Create a template
     * @request POST:/api/templates
     * @secure
     */
    createTemplate: (data: CreateTemplateReq, params: RequestParams = {}) =>
      this.request<CreateTemplateData, ErrorResponse>({
        path: `/api/templates`,
        method: "POST",
        body: data,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Templates
     * @name AllTemplates
     * @summary Get all templates
     * @request GET:/api/templates/all
     * @secure
     */
    allTemplates: (params: RequestParams = {}) =>
      this.request<AllTemplatesData, ErrorResponse>({
        path: `/api/templates/all`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),
  };
  runner = {
    /**
     * @description Uses the sandbox template permissions and limits. Missing or unsupported lang selects the first configured language. Unknown sandbox UIDs redirect to /.
     *
     * @tags Runner
     * @name ExecuteCode
     * @summary Execute code in a sandbox
     * @request POST:/r/{uid}/execute
     */
    executeCode: (
      { uid }: ExecuteCodeParams,
      data: ExecuteCodePayload,
      params: RequestParams = {},
    ) =>
      this.request<ExecuteCodeData, void | ErrorResponse>({
        path: `/r/${uid}/execute`,
        method: "POST",
        body: data,
        type: ContentType.UrlEncoded,
        format: "json",
        ...params,
      }),
  };
}
