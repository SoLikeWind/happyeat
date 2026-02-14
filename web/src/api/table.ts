import { api } from './client'
import type { Table, TableCategory } from './types'

export async function listTableCategories(params?: { current?: number; pageSize?: number; name?: string }) {
  const { data } = await api.get<{ categories: TableCategory[]; total: number }>('/central/v1/table/categories', { params })
  return data
}

export async function createTableCategory(body: { name: string; description?: string }) {
  await api.post('/central/v1/table/category', body)
}

export async function getTableCategory(id: number) {
  const { data } = await api.get<{ category: TableCategory }>(`/central/v1/table/category/${id}`)
  return data
}

export async function updateTableCategory(id: number, category: TableCategory) {
  await api.put(`/central/v1/table/category/${id}`, { category })
}

export async function deleteTableCategory(id: number) {
  await api.delete(`/central/v1/table/category/${id}`)
}

export async function listTables(params?: { current?: number; pageSize?: number; code?: string; status?: string; category?: string }) {
  const { data } = await api.get<{ tables: Table[]; total: number }>('/central/v1/tables', { params })
  return data
}

export async function createTable(body: { code: string; status: string; capacity: number; category_id: number; qr_code?: string }) {
  await api.post('/central/v1/tables', body)
}

export async function getTable(id: number) {
  const { data } = await api.get<{ table: Table }>(`/central/v1/table/${id}`)
  return data
}

export async function updateTable(id: number, table: Table) {
  await api.put(`/central/v1/table/${id}`, { table })
}

export async function deleteTable(id: number) {
  await api.delete(`/central/v1/table/${id}`)
}
