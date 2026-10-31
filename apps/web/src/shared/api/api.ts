import { request, type RequestOptions } from './client'
import type { paths } from './types'

type Method = 'get' | 'post' | 'put' | 'patch' | 'delete'
type Prefix = '/api/v1'

type RelPath<K> = K extends `${Prefix}${infer R}` ? R : never
type PathsWith<M extends Method> = {
  [K in keyof paths]: [NonNullable<paths[K][M]>] extends [never] ? never : RelPath<K>
}[keyof paths]

type Op<P extends string, M extends Method> = NonNullable<paths[`${Prefix}${P}` & keyof paths][M]>

type JsonOf<T> = T extends { content: { 'application/json': infer J } } ? J : never
type OkBody<O> = O extends { responses: infer R }
  ? JsonOf<R[Extract<keyof R, 200 | 201 | 202>]>
  : never
type Result<O> = [OkBody<O>] extends [never] ? void : OkBody<O>

type QueryOf<O> = O extends { parameters: { query?: infer Q } } ? NonNullable<Q> : never
type BodyOf<O> = O extends { requestBody?: infer B }
  ?
      | JsonOf<NonNullable<B>>
      | (NonNullable<B> extends { content: { 'multipart/form-data': unknown } } ? FormData : never)
  : never

type PathParams<P extends string> = P extends `${string}{${infer Name}}${infer Rest}`
  ? Name | PathParams<Rest>
  : never

type CallOptions<P extends string, O> = Omit<RequestOptions, 'body' | 'query' | 'method'> &
  ([QueryOf<O>] extends [never] ? { query?: undefined } : { query?: QueryOf<O> }) &
  ([BodyOf<O>] extends [never] ? { body?: undefined } : { body?: BodyOf<O> }) &
  ([PathParams<P>] extends [never]
    ? { params?: undefined }
    : { params: Record<PathParams<P>, string> })

function call(method: Method, path: string, options: Record<string, unknown> = {}) {
  const { params, query, body, ...rest } = options as {
    params?: Record<string, string>
    query?: RequestOptions['query']
    body?: unknown
  } & RequestOptions
  const url = path.replace(/\{(\w+)\}/g, (_, name: string) =>
    encodeURIComponent(params?.[name] ?? ''),
  )
  return request(url, { ...rest, method: method.toUpperCase(), query, body })
}

const make =
  <M extends Method>(method: M) =>
  <P extends PathsWith<M>>(
    path: P,
    ...args: [PathParams<P>] extends [never]
      ? [options?: CallOptions<P, Op<P, M>>]
      : [options: CallOptions<P, Op<P, M>>]
  ): Promise<Result<Op<P, M>>> =>
    call(method, path, args[0] as Record<string, unknown>) as Promise<Result<Op<P, M>>>
export const api = {
  get: make('get'),
  post: make('post'),
  put: make('put'),
  patch: make('patch'),
  delete: make('delete'),
}
