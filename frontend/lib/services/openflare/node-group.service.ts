import { OpenFlareBaseService } from './base.service';

export interface NodeGroupItem {
  id: number;
  name: string;
  monthly_bytes_limit: number;
  nodes: { id: number; node_id: string; name: string }[];
}

export interface NodeGroupPayload {
  name: string;
  monthly_bytes_limit: number;
  node_ids: number[];
}

export class NodeGroupService extends OpenFlareBaseService {
  protected static override readonly basePath = '/api/v1/d/node-groups';

  static list(): Promise<NodeGroupItem[]> {
    return this.get('');
  }

  static create(payload: NodeGroupPayload): Promise<NodeGroupItem> {
    return this.post('', payload);
  }

  static update(id: number, payload: NodeGroupPayload): Promise<NodeGroupItem> {
    return this.put(`/${id}`, payload);
  }

  static remove(id: number): Promise<void> {
    return super.delete(`/${id}`);
  }
}
