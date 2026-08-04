import { BaseService } from '@/lib/services/core';
import type {
  ChannelInput,
  PaymentChannel,
  PaymentOrder,
  PlanInput,
  RedeemCode,
  ResourceDomain,
  ResourceOrigin,
  ResourceRoute,
  ResourceZone,
  SubscriptionPlan,
  UserSubscription,
} from './types';

export class CustomService extends BaseService {
  protected static override readonly basePath = '/api/v1/custom';
  static listPlans(): Promise<SubscriptionPlan[]> {
    return this.get('/plans');
  }
  static listChannels(): Promise<PaymentChannel[]> {
    return this.get('/payment/channels');
  }
  static getSubscription(): Promise<UserSubscription> {
    return this.get('/subscription');
  }
  static listOrders(): Promise<PaymentOrder[]> {
    return this.get('/orders');
  }
  static redeem(code: string): Promise<UserSubscription> {
    return this.post('/redeem', { code });
  }
  static createOrder(
    planId: number,
    channelId?: number,
  ): Promise<{ order: PaymentOrder | null; payment_url: string }> {
    return this.post('/orders', {
      plan_id: planId,
      channel_id: channelId ?? 0,
    });
  }
  static listZones(): Promise<ResourceZone[]> {
    return this.get('/resources/zones');
  }
  static getZone(
    id: number,
  ): Promise<{ zone: ResourceZone; domains: ResourceDomain[] }> {
    return this.get(`/resources/zones/${id}`);
  }
  static createZone(domain: string): Promise<ResourceZone> {
    return this.post('/resources/zones', { domain });
  }
  static deleteZone(id: number): Promise<void> {
    return this.post(`/resources/zones/${id}/delete`);
  }
  static createDomain(zoneId: number, domain: string): Promise<ResourceDomain> {
    return this.post(`/resources/zones/${zoneId}/domains`, {
      domain,
      cert_id: null,
    });
  }
  static listOrigins(): Promise<ResourceOrigin[]> {
    return this.get('/resources/origins');
  }
  static createOrigin(payload: {
    name: string;
    address: string;
    remark: string;
  }): Promise<ResourceOrigin> {
    return this.post('/resources/origins', payload);
  }
  static deleteOrigin(id: number): Promise<void> {
    return this.post(`/resources/origins/${id}/delete`);
  }
  static listRoutes(): Promise<ResourceRoute[]> {
    return this.get('/resources/proxy-routes');
  }
  static createRoute(payload: Record<string, unknown>): Promise<ResourceRoute> {
    return this.post('/resources/proxy-routes', payload);
  }
  static deleteRoute(id: number): Promise<void> {
    return this.post(`/resources/proxy-routes/${id}/delete`);
  }
  static publish(): Promise<{ version: string; checksum: string }> {
    return this.post('/resources/publish');
  }
  static adminListPlans(): Promise<SubscriptionPlan[]> {
    return this.get('/admin/plans');
  }
  static adminCreatePlan(payload: PlanInput): Promise<SubscriptionPlan> {
    return this.post('/admin/plans', payload);
  }
  static adminUpdatePlan(
    id: number,
    payload: PlanInput,
  ): Promise<SubscriptionPlan> {
    return this.put('/admin/plans/' + id, payload);
  }
  static adminDeletePlan(id: number): Promise<void> {
    return this.delete(`/admin/plans/${id}`);
  }
  static adminListRedeemCodes(): Promise<RedeemCode[]> {
    return this.get('/admin/redeem-codes');
  }
  static adminCreateRedeemCode(planId: number): Promise<RedeemCode> {
    return this.post('/admin/redeem-codes', { plan_id: planId });
  }
  static adminListChannels(): Promise<PaymentChannel[]> {
    return this.get('/admin/payment/channels');
  }
  static adminCreateChannel(payload: ChannelInput): Promise<PaymentChannel> {
    return this.post('/admin/payment/channels', payload);
  }
  static adminUpdateChannel(
    id: number,
    payload: ChannelInput,
  ): Promise<PaymentChannel> {
    return this.put('/admin/payment/channels/' + id, payload);
  }
  static adminDeleteChannel(id: number): Promise<void> {
    return this.delete(`/admin/payment/channels/${id}`);
  }
}
