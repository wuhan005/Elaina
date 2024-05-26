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
    dns: { [key: string]: string };
    maxContainer: number;
    maxContainerPerIp: number;
}

export interface ListTemplatesResp {
    templates: Template[];
    total: number;
}

export function listTemplates(params: { page: number; pageSize: number }) {
    return axios.get<ListTemplatesResp, ListTemplatesResp>('/api/templates', {params});
}

export function allTemplates() {
    return axios.get<Template[], Template[]>('/api/templates/all');
}

export interface CreateTemplateReq {
    name: string;
    language: string[];
    timeout: number;
    maxCpus: number;
    maxMemory: number;
    internetAccess: boolean;
    dns: { [key: string]: string };
    maxContainer: number;
    maxContainerPerIp: number;
}

export function createTemplate(data: CreateTemplateReq) {
    return axios.post<Template, Template>('/api/templates', data);
}

export function getTemplate(id: string) {
    return axios.get<Template, Template>(`/api/template/${id}`);
}

export interface UpdateTemplateReq {
    name: string;
    language: string[];
    timeout: number;
    maxCpus: number;
    maxMemory: number;
    internetAccess: boolean;
    dns: { [key: string]: string };
    maxContainer: number;
    maxContainerPerIp: number;
}

export function updateTemplate(id: string, data: UpdateTemplateReq) {
    return axios.put<string, string>(`/api/template/${id}`, data);
}

export function deleteTemplate(id: number) {
    return axios.delete<string, string>(`/api/template/${id}`);
}
