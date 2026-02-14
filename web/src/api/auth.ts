import { api } from './client'
import type { LoginReq, LoginReply } from './types'

export async function login(data: LoginReq): Promise<LoginReply> {
  const res = await api.post<LoginReply>('/central/v1/auth/login', data)
  return res.data
}
