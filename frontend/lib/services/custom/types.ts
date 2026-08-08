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
export interface RedeemCode {
  id: number;
  code: string;
  plan_id: number;
  status: 'unused' | 'used';
  used_by?: number | string | null;
  used_at?: string | null;
  plan?: SubscriptionPlan;
}
export interface ResourceZone {
  id: number;
  domain: string;
  domain_count?: number;
  created_at: string;
  updated_at: string;
  claims_ownership?: boolean;
  verification_status?: string;
  verification_token?: string;
}
export interface ResourceSite {
  zone: ResourceZone;
  domain: ResourceDomain;
}
export interface ResourceDomain {
  id: number;
  zone_id: number;
  domain: string;
  proxy_route_id?: number | null;
  verification_status?: string;
  verification_token?: string;
  cert_id?: number | null;
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
  upstream_weight_list?: number[];
  enabled: boolean;
  upstream_type: string;
  enable_https?: boolean;
  redirect_http?: boolean;
  limit_rate?: string;
  limit_req_per_ip?: string;
  cache_enabled?: boolean;
  cache_policy?: string;
  cache_rule_list?: string[];
  origin_host?: string;
  limit_conn_per_server?: number;
  limit_conn_per_ip?: number;
  basic_auth_enabled?: boolean;
  basic_auth_username?: string;
  pages_project_id?: number | null;
}

export interface ResourceCertificate {
  id: number;
  name: string;
  remark?: string;
  provider?: string;
  primary_domain?: string;
  other_domains?: string;
  expires_at?: string | null;
}

export interface ResourcePagesProject {
  id: number;
  name: string;
  slug: string;
  enabled: boolean;
  active_deployment_id?: number | null;
}

export interface DnsAccountItem {
  id: number;
  name: string;
  type: string;
  created_at: string;
  updated_at: string;
}

export interface ResourcePolicy {
  cname: string;
  global_rules: ResourceWafRule[];
  default_limit_rate: string;
  default_limit_req_per_ip: string;
}

export interface ResourceWafRule {
  id: number;
  name: string;
  host: string;
  enabled: boolean;
  is_global: boolean;
  applied_site_count?: number;
}

export interface ResourceRouteWaf {
  route_id: number;
  global_rule_group?: ResourceWafRule | null;
  rule_groups: ResourceWafRule[];
  applied_rule_groups: ResourceWafRule[];
  applied_ids: number[];
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
