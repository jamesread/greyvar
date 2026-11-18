import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { DatEditorService } from '../../gen/deditor/v1/deditor_pb.ts'

const transport = createConnectTransport({
  baseUrl: '',
})

export const client = createClient(DatEditorService, transport)

export function textureUrl (kind, filename) {
  return `/res/img/textures/${kind}/${encodeURIComponent(filename)}`
}
