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

export interface ChangePasswordForm {
  currentPassword: string;
  /** Password must contain 8-72 UTF-8 bytes. */
  newPassword: string;
}

export interface ConfirmPasswordForm {
  /** The current password of the signed-in user. */
  password: string;
}

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

export interface CreateUserReq {
  /** Password must contain 8-72 UTF-8 bytes. */
  password: string;
  /** 2-32 letters, digits, '_', '.' or '-', starting with a letter or digit. */
  username: string;
}

export interface EnableTwoFactorForm {
  /** A 6-digit code from the authenticator app. */
  code: string;
}

export interface ErrorResponse {
  msg: string;
}

export interface ExecutionDailyCount {
  /** Local day in YYYY-MM-DD. */
  date: string;
  failed: number;
  total: number;
}

export interface ExecutionDetail {
  clientAddress: string;
  /** Empty for rejected requests. */
  code: string;
  /** Size of the submitted source code in bytes. */
  codeBytes: number;
  createdAt: string;
  /** WASM run time in nanoseconds; 0 when the task did not run. */
  duration: number;
  error: string;
  id: number;
  language: string;
  sandbox?: ExecutionSandbox | null;
  sandboxID: number;
  status: "success" | "failed" | "timeout" | "rejected" | "error";
  stderr: string;
  stdout: string;
  /** True when code or output was cut to 64 KiB. */
  truncated: boolean;
  userAgent: string;
}

export interface ExecutionResult {
  end_at: number;
  result: RuntimeExecOutput;
  start_at: number;
}

export interface ExecutionSandbox {
  deleted: boolean;
  id: number;
  name: string;
  uid: string;
}

export interface ExecutionStats {
  /** One entry per day, oldest first. */
  daily: ExecutionDailyCount[];
  failed24h: number;
  last24h: number;
  total: number;
}

export interface ExecutionSummary {
  clientAddress: string;
  /** Size of the submitted source code in bytes. */
  codeBytes: number;
  createdAt: string;
  /** WASM run time in nanoseconds; 0 when the task did not run. */
  duration: number;
  error: string;
  id: number;
  language: string;
  sandbox?: ExecutionSandbox | null;
  sandboxID: number;
  status: "success" | "failed" | "timeout" | "rejected" | "error";
  userAgent: string;
}

export interface ListExecutionsResp {
  executions: ExecutionSummary[];
  total: number;
}

export interface ListSandboxesResp {
  sandboxes: Sandbox[];
  total: number;
}

export interface ListTemplatesResp {
  templates: Template[];
  total: number;
}

export interface RecoveryCodesResp {
  /** One-time recovery codes. They are only shown once. */
  codes: string[];
}

export interface ResetPasswordReq {
  /** Password must contain 8-72 UTF-8 bytes. */
  password: string;
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
  username: string;
}

export interface SignInResp {
  /** Session token for the Authorization header ("Token <token>"). */
  token: string;
  /** When true, the token can only be used to submit a verification code to signInTwoFactor. */
  twoFactorRequired: boolean;
}

export interface SignInTwoFactorForm {
  /** A 6-digit TOTP code or an unused recovery code. */
  code: string;
}

export interface Template {
  createdAt: string;
  id: number;
  internetAccess: boolean;
  language?: string[] | null;
  maxMemory: number;
  name: string;
  timeout: number;
  updatedAt: string;
}

export interface TwoFactorSetupResp {
  /** PNG QR code as a data URL. */
  qrCode: string;
  /** Base32 TOTP secret for manual entry. */
  secret: string;
  /** otpauth:// URL encoded in the QR code. */
  url: string;
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

export interface User {
  createdAt: string;
  id: number;
  lastSignInAt?: string | null;
  recoveryCodesRemaining: number;
  twoFactorEnabled: boolean;
  updatedAt: string;
  username: string;
}

export interface DtoResponseExecutionDetail {
  data: ExecutionDetail;
}

export interface DtoResponseExecutionResult {
  data: ExecutionResult;
}

export interface DtoResponseExecutionStats {
  data: ExecutionStats;
}

export interface DtoResponseListExecutionsResp {
  data: ListExecutionsResp;
}

export interface DtoResponseListSandboxesResp {
  data: ListSandboxesResp;
}

export interface DtoResponseListTemplatesResp {
  data: ListTemplatesResp;
}

export interface DtoResponseRecoveryCodesResp {
  data: RecoveryCodesResp;
}

export interface DtoResponseSandbox {
  data: Sandbox;
}

export interface DtoResponseSignInResp {
  data: SignInResp;
}

export interface DtoResponseTemplate {
  data: Template;
}

export interface DtoResponseTwoFactorSetupResp {
  data: TwoFactorSetupResp;
}

export interface DtoResponseUser {
  data: User;
}

export interface DtoResponseArrayTemplate {
  data: Template[];
}

export interface DtoResponseArrayUser {
  data: User[];
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

export type DisableTwoFactorData = DtoResponseString;

export type EnableTwoFactorData = DtoResponseRecoveryCodesResp;

export type RegenerateRecoveryCodesData = DtoResponseRecoveryCodesResp;

export type SetupTwoFactorData = DtoResponseTwoFactorSetupResp;

export type ChangePasswordData = DtoResponseString;

export type ProfileData = DtoResponseUser;

export type SignInData = DtoResponseSignInResp;

export type SignInTwoFactorData = DtoResponseSignInResp;

export type SignOutData = DtoResponseString;

export interface GetExecutionParams {
  /** ID */
  id: number;
}

export type GetExecutionData = DtoResponseExecutionDetail;

export interface ListExecutionsParams {
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
  /** Only executions of this sandbox */
  sandboxID?: number;
  /** Only executions with this status */
  status?: "success" | "failed" | "timeout" | "rejected" | "error";
}

export type ListExecutionsData = DtoResponseListExecutionsResp;

export interface ExecutionStatsParams {
  /**
   * Number of days in daily, including today (1-90)
   * @default 14
   */
  days?: number;
  /**
   * IANA time zone used to split days
   * @default "UTC"
   */
  timezone?: string;
}

export type ExecutionStatsData = DtoResponseExecutionStats;

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

export interface DeleteUserParams {
  /** ID */
  id: number;
}

export type DeleteUserData = DtoResponseString;

export interface ResetUserTwoFactorParams {
  /** ID */
  id: number;
}

export type ResetUserTwoFactorData = DtoResponseString;

export interface ResetUserPasswordParams {
  /** ID */
  id: number;
}

export type ResetUserPasswordData = DtoResponseString;

export type ListUsersData = DtoResponseArrayUser;

export type CreateUserData = DtoResponseUser;

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
 * Management and WASM code execution API. Successful JSON responses use a data envelope; errors contain msg. All HTTP bodies are capped by app.wasm_limits.max_request_bytes; oversized bodies return HTTP 413. Buffered bodies share app.wasm_limits.max_buffered_request_bytes; requests that exceed it return HTTP 503.
 */
export class Api<
  SecurityDataType extends unknown,
> extends HttpClient<SecurityDataType> {
  account = {
    /**
     * No description
     *
     * @tags Account
     * @name DisableTwoFactor
     * @summary Disable two-factor authentication
     * @request POST:/api/account/2fa/disable
     * @secure
     */
    disableTwoFactor: (data: ConfirmPasswordForm, params: RequestParams = {}) =>
      this.request<DisableTwoFactorData, ErrorResponse>({
        path: `/api/account/2fa/disable`,
        method: "POST",
        body: data,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Confirms the secret from setupTwoFactor and returns new recovery codes.
     *
     * @tags Account
     * @name EnableTwoFactor
     * @summary Enable two-factor authentication
     * @request POST:/api/account/2fa/enable
     * @secure
     */
    enableTwoFactor: (data: EnableTwoFactorForm, params: RequestParams = {}) =>
      this.request<EnableTwoFactorData, ErrorResponse>({
        path: `/api/account/2fa/enable`,
        method: "POST",
        body: data,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Replaces all existing recovery codes.
     *
     * @tags Account
     * @name RegenerateRecoveryCodes
     * @summary Regenerate recovery codes
     * @request POST:/api/account/2fa/recovery-codes
     * @secure
     */
    regenerateRecoveryCodes: (
      data: ConfirmPasswordForm,
      params: RequestParams = {},
    ) =>
      this.request<RegenerateRecoveryCodesData, ErrorResponse>({
        path: `/api/account/2fa/recovery-codes`,
        method: "POST",
        body: data,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Generates a new TOTP secret. Two-factor authentication stays disabled until enableTwoFactor confirms a code.
     *
     * @tags Account
     * @name SetupTwoFactor
     * @summary Start two-factor authentication setup
     * @request POST:/api/account/2fa/setup
     * @secure
     */
    setupTwoFactor: (params: RequestParams = {}) =>
      this.request<SetupTwoFactorData, ErrorResponse>({
        path: `/api/account/2fa/setup`,
        method: "POST",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Other sessions of the user are signed out; the current session stays valid.
     *
     * @tags Account
     * @name ChangePassword
     * @summary Change the password of the signed-in user
     * @request PUT:/api/account/password
     * @secure
     */
    changePassword: (data: ChangePasswordForm, params: RequestParams = {}) =>
      this.request<ChangePasswordData, ErrorResponse>({
        path: `/api/account/password`,
        method: "PUT",
        body: data,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
  auth = {
    /**
     * No description
     *
     * @tags Auth
     * @name Profile
     * @summary Get the signed-in user
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
     * @description When the user has two-factor authentication enabled, twoFactorRequired is true and the returned token must be sent with signInTwoFactor within 5 minutes.
     *
     * @tags Auth
     * @name SignIn
     * @summary Sign in with a username and password
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
     * @description Accepts a TOTP code or an unused recovery code. After 5 wrong codes or 5 minutes the pending sign-in is discarded and the client must sign in again.
     *
     * @tags Auth
     * @name SignInTwoFactor
     * @summary Complete sign-in with a verification code
     * @request POST:/api/auth/sign-in/2fa
     * @secure
     */
    signInTwoFactor: (data: SignInTwoFactorForm, params: RequestParams = {}) =>
      this.request<SignInTwoFactorData, ErrorResponse>({
        path: `/api/auth/sign-in/2fa`,
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
  executions = {
    /**
     * No description
     *
     * @tags Executions
     * @name GetExecution
     * @summary Get a sandbox execution
     * @request GET:/api/execution/{id}
     * @secure
     */
    getExecution: ({ id }: GetExecutionParams, params: RequestParams = {}) =>
      this.request<GetExecutionData, ErrorResponse>({
        path: `/api/execution/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Newest first. Code and output are only included in getExecution.
     *
     * @tags Executions
     * @name ListExecutions
     * @summary List sandbox executions
     * @request GET:/api/executions
     * @secure
     */
    listExecutions: (
      query: ListExecutionsParams = {},
      params: RequestParams = {},
    ) =>
      this.request<ListExecutionsData, ErrorResponse>({
        path: `/api/executions`,
        method: "GET",
        query: query,
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Executions
     * @name ExecutionStats
     * @summary Get execution statistics
     * @request GET:/api/executions/stats
     * @secure
     */
    executionStats: (
      query: ExecutionStatsParams = {},
      params: RequestParams = {},
    ) =>
      this.request<ExecutionStatsData, ErrorResponse>({
        path: `/api/executions/stats`,
        method: "GET",
        query: query,
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
     * @description Zero timeout uses 10 seconds. Zero memory inherits app.wasm_limits.max_memory_mib; nonzero memory is capped by that ceiling for each stage.
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
     * @description Zero timeout uses 10 seconds. Zero memory inherits app.wasm_limits.max_memory_mib; nonzero memory is capped by that ceiling for each stage.
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
  users = {
    /**
     * No description
     *
     * @tags Users
     * @name DeleteUser
     * @summary Delete a user
     * @request DELETE:/api/user/{id}
     * @secure
     */
    deleteUser: ({ id }: DeleteUserParams, params: RequestParams = {}) =>
      this.request<DeleteUserData, ErrorResponse>({
        path: `/api/user/${id}`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description For users who lost their authenticator and recovery codes. Signs the user out of all sessions.
     *
     * @tags Users
     * @name ResetUserTwoFactor
     * @summary Turn off two-factor authentication for a user
     * @request DELETE:/api/user/{id}/2fa
     * @secure
     */
    resetUserTwoFactor: (
      { id }: ResetUserTwoFactorParams,
      params: RequestParams = {},
    ) =>
      this.request<ResetUserTwoFactorData, ErrorResponse>({
        path: `/api/user/${id}/2fa`,
        method: "DELETE",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Signs the user out of all sessions.
     *
     * @tags Users
     * @name ResetUserPassword
     * @summary Set a new password for a user
     * @request PUT:/api/user/{id}/password
     * @secure
     */
    resetUserPassword: (
      { id }: ResetUserPasswordParams,
      data: ResetPasswordReq,
      params: RequestParams = {},
    ) =>
      this.request<ResetUserPasswordData, ErrorResponse>({
        path: `/api/user/${id}/password`,
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
     * @tags Users
     * @name ListUsers
     * @summary List users
     * @request GET:/api/users
     * @secure
     */
    listUsers: (params: RequestParams = {}) =>
      this.request<ListUsersData, ErrorResponse>({
        path: `/api/users`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * No description
     *
     * @tags Users
     * @name CreateUser
     * @summary Create a user
     * @request POST:/api/users
     * @secure
     */
    createUser: (data: CreateUserReq, params: RequestParams = {}) =>
      this.request<CreateUserData, ErrorResponse>({
        path: `/api/users`,
        method: "POST",
        body: data,
        secure: true,
        type: ContentType.Json,
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
