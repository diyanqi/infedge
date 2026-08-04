export interface SubscriptionPlan {
  id: number;
  name: string;
  description: string;
  price_fen: number;
  billing_months: number;
  high_speed_bytes: number;
  throttle_bytes_per_sec: number;
  daily_publish_limit: number;
  max_zones: number;
  max_origins: number;
  max_routes: number;
  max_pages: number;
  enabled: boolean;
}
export interface PaymentChannel {
  id: number;
  name: string;
  gateway: string;
  pid: string;
  enabled: boolean;
  sort: number;
}
export interface UserSubscription {
  id: number;
  user_id: string;
  plan_id: number;
  status: string;
  starts_at: string;
  expires_at: string;
  plan?: SubscriptionPlan;
}
export interface PaymentOrder {
  id: number;
  order_no: string;
  plan_id: number;
  channel_id: number;
  amount_fen: number;
  status: string;
  created_at: string;
  paid_at?: string | null;
}
export interface ResourceZone {
  id: number;
  domain: string;
  domain_count?: number;
  created_at: string;
  updated_at: string;
}
export interface ResourceDomain {
  id: number;
  zone_id: number;
  domain: string;
  proxy_route_id?: number | null;
}
export interface ResourceOrigin {
  id: number;
  name: string;
  address: string;
  remark: string;
  owner_id?: string;
}
export interface ResourceRoute {
  id: number;
  site_name: string;
  zone_domain_ids: number[];
  zone_domains: ResourceDomain[];
  origin_id: number | null;
  origin_url: string;
  upstream_list: string[];
  enabled: boolean;
  upstream_type: string;
}
export interface ResourceNodeGroup {
  id: number;
  name: string;
  monthly_bytes_limit: number;
  nodes: { id: number; node_id: string; name: string }[];
}
export interface PlanInput {
  name: string;
  description: string;
  price_fen: number;
  billing_months: number;
  high_speed_bytes: number;
  throttle_bytes_per_sec: number;
  daily_publish_limit: number;
  max_zones: number;
  max_origins: number;
  max_routes: number;
  max_pages: number;
  enabled: boolean;
}
export interface ChannelInput {
  name: string;
  gateway: string;
  pid: string;
  secret_key: string;
  enabled: boolean;
  sort: number;
}
