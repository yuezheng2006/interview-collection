import { defineStore } from 'pinia'
import { listDocuments, createDocument as apiCreate, updateDocument as apiUpdate, deleteDocument as apiDelete, renameDocument as apiRename } from '@/services/document'

export type DocumentItem = {
  id: string
  title: string
  content: string
  createdAt: string
  updatedAt: string
  wordCount: number
}

export const useDocumentStore = defineStore('document', {
  state: () => ({
    documents: [] as DocumentItem[],
    currentDocument: null as DocumentItem | null,
    isLoading: false,
  }),
  actions: {
    async fetchDocuments() {
      this.isLoading = true
      try {
        const res = await listDocuments()
        this.documents = res.data as unknown as DocumentItem[]
      } finally {
        this.isLoading = false
      }
    },
    async createDocument(title: string) {
      const res = await apiCreate(title)
      const doc = res.data as unknown as DocumentItem
      this.documents.unshift(doc)
      this.currentDocument = doc
    },
    async renameDocument(id: string, title: string) {
      const res = await apiRename(id, title)
      const doc = res.data as unknown as DocumentItem
      const idx = this.documents.findIndex(d => d.id === id)
      if (idx >= 0) this.documents[idx] = doc
      if (this.currentDocument?.id === id) this.currentDocument = doc
      return res
    },
    async saveDocumentContent(content: string) {
      if (!this.currentDocument) return
      const res = await apiUpdate(this.currentDocument.id, { content })
      const updated = res.data as unknown as DocumentItem
      this.currentDocument = updated
      const idx = this.documents.findIndex(d => d.id === updated.id)
      if (idx >= 0) this.documents[idx] = updated
    },
    async deleteDocument(id: string) {
      await apiDelete(id)
      this.documents = this.documents.filter(d => d.id !== id)
      if (this.currentDocument?.id === id) this.currentDocument = null
    }
  },
})

