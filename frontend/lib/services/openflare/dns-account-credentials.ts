export type DnsAccountProviderType =
  | 'cloudflare'
  | 'aliyun'
  | 'tencent'
  | 'huawei';

export interface DnsAccountCredentialField {
  key: string;
  label: string;
  placeholder: string;
  required: boolean;
  type?: 'text' | 'password';
  help?: string;
}

export interface DnsAccountCredentialConfig {
  provider: DnsAccountProviderType;
  label: string;
  fields: DnsAccountCredentialField[];
}

export const DNS_ACCOUNT_CREDENTIALS: Record<
  DnsAccountProviderType,
  DnsAccountCredentialConfig
> = {
  cloudflare: {
    provider: 'cloudflare',
    label: 'Cloudflare',
    fields: [
      {
        key: 'api_token',
        label: 'API Token',
        placeholder: '粘贴 Cloudflare API Token',
        required: true,
        type: 'password',
      },
    ],
  },
  aliyun: {
    provider: 'aliyun',
    label: '阿里云 DNS',
    fields: [
      {
        key: 'access_key',
        label: 'AccessKey ID',
        placeholder: '粘贴阿里云 AccessKey ID',
        required: true,
        type: 'password',
      },
      {
        key: 'access_key_secret',
        label: 'AccessKey Secret',
        placeholder: '粘贴阿里云 AccessKey Secret',
        required: true,
        type: 'password',
      },
      {
        key: 'security_token',
        label: 'Security Token（可选）',
        placeholder: '使用临时 STS 凭据时填写',
        required: false,
        type: 'password',
      },
      {
        key: 'region_id',
        label: 'Region ID（可选）',
        placeholder: '默认 cn-hangzhou，如 cn-beijing',
        required: false,
      },
    ],
  },
  tencent: {
    provider: 'tencent',
    label: '腾讯云 DNSPod',
    fields: [
      {
        key: 'secret_id',
        label: 'SecretId',
        placeholder: '粘贴腾讯云 SecretId',
        required: true,
        type: 'password',
      },
      {
        key: 'secret_key',
        label: 'SecretKey',
        placeholder: '粘贴腾讯云 SecretKey',
        required: true,
        type: 'password',
      },
      {
        key: 'session_token',
        label: 'SessionToken（可选）',
        placeholder: '使用临时密钥时填写',
        required: false,
        type: 'password',
      },
      {
        key: 'region',
        label: 'Region（可选）',
        placeholder: '例如 ap-guangzhou',
        required: false,
      },
    ],
  },
  huawei: {
    provider: 'huawei',
    label: '华为云 DNS',
    fields: [
      {
        key: 'access_key_id',
        label: 'Access Key ID（AK）',
        placeholder: '粘贴华为云 AK',
        required: true,
        type: 'password',
      },
      {
        key: 'secret_access_key',
        label: 'Secret Access Key（SK）',
        placeholder: '粘贴华为云 SK',
        required: true,
        type: 'password',
      },
      {
        key: 'region',
        label: 'Region',
        placeholder: '例如 cn-north-4',
        required: true,
        help: '华为云 DNS 服务支持的区域，如 cn-north-4、cn-east-3',
      },
    ],
  },
};

export function buildDnsAccountAuthorization(
  type: string,
  credentials: Record<string, string>,
): string {
  const config = DNS_ACCOUNT_CREDENTIALS[type as DnsAccountProviderType];
  const payload: Record<string, string> = {};
  for (const field of config?.fields ?? []) {
    const value = credentials[field.key]?.trim();
    if (value) {
      payload[field.key] = value;
    }
  }
  return JSON.stringify(payload);
}
