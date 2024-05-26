import axios from 'axios';
import {Template} from "@/api/template";

export interface Sandbox {
    id: number;
    createdAt: Date;
    updatedAt: Date;

    uid: string;
    name: string;
    templateID: number;
    template: Template;
    placeholder: string;
    editable: boolean;
}

export interface ListSandboxesResp {
    sandboxes: Sandbox[];
    total: number;
}

export function listSandboxes(params: { page: number; pageSize: number }) {
    return axios.get<ListSandboxesResp, ListSandboxesResp>('/api/sandboxes', {params});
}

export interface CreateSandboxReq {
    name: string;
    templateID: number;
    placeholder: string;
    editable: boolean;
}

export function createSandbox(data: CreateSandboxReq) {
    return axios.post<Sandbox, Sandbox>('/api/sandboxes', data);
}

export function getSandbox(id: string) {
    return axios.get<Sandbox, Sandbox>(`/api/sandbox/${id}`);
}

export interface UpdateSandboxReq {
    name: string;
    templateID: number;
    placeholder: string;
    editable: boolean;
}

export function updateSandbox(id: string, data: UpdateSandboxReq) {
    return axios.put<string, string>(`/api/sandbox/${id}`, data);
}

export function deleteSandbox(id: number) {
    return axios.delete<string, string>(`/api/sandbox/${id}`);
}
