import { BaseService } from '@/lib/services/core';
import type {
  AcmeAccountItem,
  DnsAccountItem,
  DnsAccountMutationPayload,
  TlsCertificateApplyPayload,
  TlsCertificateContentItem,
  TlsCertificateDetailItem,
  TlsCertificateFileImportPayload,
  TlsCertificateItem,
  TlsCertificateMutationPayload,
} from '@/lib/services/openflare';

export interface TlsCertificateApi {
  list(): Promise<TlsCertificateItem[]>;
  getById(id: number): Promise<TlsCertificateDetailItem>;
  getContent(id: number): Promise<TlsCertificateContentItem>;
  create(payload: TlsCertificateMutationPayload): Promise<TlsCertificateItem>;
  update(
    id: number,
    payload: TlsCertificateMutationPayload,
  ): Promise<TlsCertificateItem>;
  deleteById(id: number): Promise<void>;
  apply(payload: TlsCertificateApplyPayload): Promise<TlsCertificateItem>;
  renew(id: number): Promise<TlsCertificateItem>;
  updateAcme(
    id: number,
    payload: TlsCertificateApplyPayload,
  ): Promise<TlsCertificateItem>;
  convertToAcme(
    id: number,
    payload: TlsCertificateApplyPayload,
  ): Promise<TlsCertificateItem>;
  importFile(
    payload: TlsCertificateFileImportPayload,
  ): Promise<TlsCertificateItem>;
  getDefaultAcmeAccount(): Promise<AcmeAccountItem>;
}

export interface DnsAccountApi {
  list(): Promise<DnsAccountItem[]>;
  create(payload: DnsAccountMutationPayload): Promise<DnsAccountItem>;
  update(
    id: number,
    payload: DnsAccountMutationPayload,
  ): Promise<DnsAccountItem>;
  deleteById(id: number): Promise<void>;
}

export class CustomTlsCertificateService extends BaseService {
  protected static override readonly basePath =
    '/api/v1/custom/resources/tls-certificates';

  static list(): Promise<TlsCertificateItem[]> {
    return this.get<TlsCertificateItem[]>('');
  }

  static getById(id: number): Promise<TlsCertificateDetailItem> {
    return this.get<TlsCertificateDetailItem>(`/${id}`);
  }

  static getContent(id: number): Promise<TlsCertificateContentItem> {
    return this.get<TlsCertificateContentItem>(`/${id}/content`);
  }

  static create(
    payload: TlsCertificateMutationPayload,
  ): Promise<TlsCertificateItem> {
    return this.post<TlsCertificateItem>('', payload);
  }

  static update(
    id: number,
    payload: TlsCertificateMutationPayload,
  ): Promise<TlsCertificateItem> {
    return this.post<TlsCertificateItem>(`/${id}/update`, payload);
  }

  static deleteById(id: number): Promise<void> {
    return this.post<void>(`/${id}/delete`);
  }

  static apply(
    payload: TlsCertificateApplyPayload,
  ): Promise<TlsCertificateItem> {
    return this.post<TlsCertificateItem>('/apply', payload);
  }

  static renew(id: number): Promise<TlsCertificateItem> {
    return this.post<TlsCertificateItem>(`/${id}/renew`);
  }

  static updateAcme(
    id: number,
    payload: TlsCertificateApplyPayload,
  ): Promise<TlsCertificateItem> {
    return this.post<TlsCertificateItem>(`/${id}/update-acme`, payload);
  }

  static convertToAcme(
    id: number,
    payload: TlsCertificateApplyPayload,
  ): Promise<TlsCertificateItem> {
    return this.post<TlsCertificateItem>(`/${id}/convert-acme`, payload);
  }

  static importFile(
    payload: TlsCertificateFileImportPayload,
  ): Promise<TlsCertificateItem> {
    const formData = new FormData();
    formData.append('name', payload.name);
    formData.append('remark', payload.remark);
    formData.append('cert_file', payload.certFile);
    formData.append('key_file', payload.keyFile);
    return this.post<TlsCertificateItem>('/import-file', formData);
  }

  static getDefaultAcmeAccount(): Promise<AcmeAccountItem> {
    return this.get<AcmeAccountItem>('/acme-account/default');
  }
}

export class CustomDnsAccountService extends BaseService {
  protected static override readonly basePath =
    '/api/v1/custom/resources/dns-accounts';

  /** 平台账号 + 我的账号（ACME 申请可选列表）。 */
  static list(): Promise<DnsAccountItem[]> {
    return this.get<DnsAccountItem[]>('/tls-certificates/dns-accounts');
  }

  /** 仅我的账号（DNS 账号管理页）。 */
  static listOwned(): Promise<DnsAccountItem[]> {
    return this.get<DnsAccountItem[]>('');
  }

  static create(payload: DnsAccountMutationPayload): Promise<DnsAccountItem> {
    return this.post<DnsAccountItem>('', payload);
  }

  static update(
    id: number,
    payload: DnsAccountMutationPayload,
  ): Promise<DnsAccountItem> {
    return this.post<DnsAccountItem>(`/${id}/update`, payload);
  }

  static deleteById(id: number): Promise<void> {
    return this.post<void>(`/${id}/delete`);
  }
}
