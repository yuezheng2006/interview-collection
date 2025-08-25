import { getHttpClient } from './api'

export async function register(username: string, password: string): Promise<{ ok: boolean }> {
  const res = await getHttpClient().post('/auth/register', { username, password })
  return res.data
}

export async function login(username: string, password: string): Promise<{ token: string }> {
  const res = await getHttpClient().post('/auth/login', { username, password })
  return res.data
}

