import { getHttpClient } from './api'
import type { AxiosResponse } from 'axios'

export type DocumentDto = {
  id: string
  title: string
  content: string
  createdAt: string
  updatedAt: string
  wordCount: number
}

export type UpdateDocumentRequest = {
  content?: string
  title?: string
}

export async function listDocuments(): Promise<AxiosResponse<DocumentDto[]>> {
  return getHttpClient().get('/documents')
}

export async function createDocument(title: string): Promise<AxiosResponse<DocumentDto>> {
  return getHttpClient().post('/documents', { title })
}

export async function updateDocument(
  id: string,
  data: UpdateDocumentRequest
): Promise<AxiosResponse<DocumentDto>> {
  return getHttpClient().put(`/documents/${id}`, data)
}

export async function renameDocument(
  id: string,
  title: string
): Promise<AxiosResponse<DocumentDto>> {
  return updateDocument(id, { title })
}

export async function deleteDocument(id: string): Promise<AxiosResponse<void>> {
  return getHttpClient().delete(`/documents/${id}`)
}

