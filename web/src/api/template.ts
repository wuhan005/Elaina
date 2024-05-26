import axios from 'axios';

export interface Template {
    id: number;
    createdAt: Date;
    updatedAt: Date;

    name: string;
    language: string[];
    timeout: number;
    maxCpus: number;
    maxMemory: number;
    internetAccess: boolean;
    dns: any;
    maxContainer: number;
    maxContainerPerIp: number;
}

export interface ListTemplatesResp {
    templates: Template[];
    total: number;
}

export function listTemplates(params: { page: number; pageSize: number }) {
    return axios.get<ListTemplatesResp>('/api/templates', {params});
}

export interface CreateTemplateReq {
    name: string;
    language: string[];
    timeout: number;
    maxCpus: number;
    maxMemory: number;
    internetAccess: boolean;
    dns: any;
    maxContainer: number;
    maxContainerPerIp: number;
}

export function createTemplate(data: CreateTemplateReq) {
    return axios.post<Template>('/api/templates', data);
}

export function getTemplate(id: number) {
    return axios.get<Template>(`/api/template/${id}`);
}

export interface UpdateTemplateReq {
    name: string;
    language: string[];
    timeout: number;
    maxCpus: number;
    maxMemory: number;
    internetAccess: boolean;
    dns: any;
    maxContainer: number;
    maxContainerPerIp: number;
}

export function updateTemplate(id: number, data: UpdateTemplateReq) {
    return axios.put<string>(`/api/template/${id}`, data);
}

export function deleteTemplate(id: number) {
    return axios.delete<string>(`/api/template/${id}`);
}
