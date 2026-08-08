'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Check, CreditCard, Package, RefreshCw, Ticket } from 'lucide-react';
import { toast } from 'sonner';
import { CustomService } from '@/lib/services/custom';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { LoadingStateWithBorder } from '@/components/layout/loading';
import { ErrorInline } from '@/components/layout/error';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

const formatBytes = (value: number) =>
  value <= 0 ? '不限' : (value / 1073741824).toFixed(2) + ' GB';
const formatSpeed = (value: number) =>
  value <= 0 ? '不额外限速' : (value / 1024).toFixed(1) + ' KB/s';

export default function PlansPage() {
  const queryClient = useQueryClient();
  const [channelID, setChannelID] = useState('');
  const [redeemCode, setRedeemCode] = useState('');
  const plans = useQuery({
    queryKey: ['custom', 'plans'],
    queryFn: () => CustomService.listPlans(),
  });
  const channels = useQuery({
    queryKey: ['custom', 'channels'],
    queryFn: () => CustomService.listChannels(),
  });
  const subscription = useQuery({
    queryKey: ['custom', 'subscription'],
    queryFn: () => CustomService.getSubscription(),
    retry: false,
  });
  const buy = useMutation({
    mutationFn: ({
      planId,
      channelId,
    }: {
      planId: number;
      channelId?: number;
    }) => CustomService.createOrder(planId, channelId),
    onSuccess: async (result) => {
      await queryClient.invalidateQueries({
        queryKey: ['custom', 'subscription'],
      });
      if (result.payment_url)
        window.open(result.payment_url, '_blank', 'noopener,noreferrer');
      toast.success(
        result.payment_url ? '订单已创建，请在新窗口完成支付' : '套餐已开通',
      );
    },
    onError: (error) =>
      toast.error(error instanceof Error ? error.message : '购买失败'),
  });
  const redeem = useMutation({
    mutationFn: (code: string) => CustomService.redeem(code),
    onSuccess: async () => {
      setRedeemCode('');
      await queryClient.invalidateQueries({
        queryKey: ['custom', 'subscription'],
      });
      toast.success('兑换成功，套餐已延长一个月');
    },
    onError: (error) =>
      toast.error(error instanceof Error ? error.message : '兑换失败'),
  });

  return (
    <div className='w-full space-y-6 py-6 px-1'>
      <div className='flex items-center justify-between gap-3'>
        <div className='flex items-center gap-2'>
          <Package className='size-5 text-primary' />
          <h1 className='text-2xl font-semibold tracking-tight'>套餐订阅</h1>
        </div>
        <Button
          variant='outline'
          size='icon'
          onClick={() => void plans.refetch()}
          disabled={plans.isFetching}
          aria-label='刷新套餐'
        >
          <RefreshCw
            className={plans.isFetching ? 'size-4 animate-spin' : 'size-4'}
          />
        </Button>
      </div>
      {subscription.data && (
        <Card className='border-primary/30'>
          <CardHeader className='pb-3'>
            <CardTitle className='text-base'>
              当前套餐：{subscription.data.plan?.name}
            </CardTitle>
            <CardDescription>
              有效期至 {new Date(subscription.data.expires_at).toLocaleString()}
            </CardDescription>
          </CardHeader>
          <CardContent className='grid gap-2 text-sm sm:grid-cols-4'>
            <span>
              高速流量：
              {formatBytes(subscription.data.plan?.high_speed_bytes ?? 0)}
            </span>
            <span>
              超额速度：
              {formatSpeed(subscription.data.plan?.throttle_bytes_per_sec ?? 0)}
            </span>
            <span>
              每日发布：{subscription.data.plan?.daily_publish_limit ?? 0} 次
            </span>
            <span>
              资源额度：域名分组 {subscription.data.plan?.max_zones ?? 0} / 规则{' '}
              {subscription.data.plan?.max_routes ?? 0}
            </span>
          </CardContent>
        </Card>
      )}
      <Card>
        <CardHeader className='pb-3'>
          <CardTitle className='flex items-center gap-2 text-base'>
            <Ticket className='size-4 text-primary' />
            套餐兑换码
          </CardTitle>
          <CardDescription>
            输入管理员提供的兑换码，可兑换对应套餐一个月。
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form
            className='flex flex-col gap-2 sm:flex-row'
            onSubmit={(event) => {
              event.preventDefault();
              const code = redeemCode.trim();
              if (code && !redeem.isPending) redeem.mutate(code);
            }}
          >
            <Input
              value={redeemCode}
              onChange={(event) => setRedeemCode(event.target.value)}
              placeholder='输入兑换码'
              maxLength={64}
              className='sm:max-w-sm'
            />
            <Button
              type='submit'
              disabled={!redeemCode.trim() || redeem.isPending}
            >
              <Ticket className='mr-2 size-4' />
              {redeem.isPending ? '兑换中...' : '立即兑换'}
            </Button>
          </form>
        </CardContent>
      </Card>
      {plans.isLoading ? (
        <LoadingStateWithBorder title='加载套餐' />
      ) : plans.isError ? (
        <ErrorInline
          message='套餐加载失败'
          onRetry={() => void plans.refetch()}
        />
      ) : (
        <div className='grid gap-4 md:grid-cols-2 xl:grid-cols-3'>
          {(plans.data ?? []).map((plan) => (
            <Card key={plan.id} className='flex flex-col border-dashed'>
              <CardHeader>
                <div className='flex items-start justify-between gap-2'>
                  <div>
                    <CardTitle>{plan.name}</CardTitle>
                    <CardDescription>
                      {plan.description || '按月提供边缘资源额度'}
                    </CardDescription>
                  </div>
                  <Badge
                    variant={plan.price_fen === 0 ? 'secondary' : 'default'}
                  >
                    {plan.price_fen === 0
                      ? '免费'
                      : '¥' + (plan.price_fen / 100).toFixed(2)}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className='flex flex-1 flex-col gap-3 text-sm'>
                <div className='space-y-2 text-muted-foreground'>
                  <p>
                    <Check className='mr-2 inline size-4 text-primary' />
                    高速流量 {formatBytes(plan.high_speed_bytes)} / 月
                  </p>
                  <p>
                    <Check className='mr-2 inline size-4 text-primary' />
                    超额后限速 {formatSpeed(plan.throttle_bytes_per_sec)}
                  </p>
                  <p>
                    <Check className='mr-2 inline size-4 text-primary' />
                    每天发布{' '}
                    {plan.daily_publish_limit <= 0
                      ? '不限'
                      : plan.daily_publish_limit}{' '}
                    次
                  </p>
                  <p>
                    <Check className='mr-2 inline size-4 text-primary' />
                    域名分组 {plan.max_zones} · 源站 {plan.max_origins} · 规则{' '}
                    {plan.max_routes}
                  </p>
                </div>
                <Button
                  className='mt-auto w-full'
                  disabled={buy.isPending}
                  onClick={() =>
                    buy.mutate({
                      planId: plan.id,
                      channelId:
                        plan.price_fen === 0
                          ? undefined
                          : Number(
                              channelID || String(channels.data?.[0]?.id ?? 0),
                            ),
                    })
                  }
                >
                  <CreditCard className='mr-2 size-4' />
                  {plan.price_fen === 0 ? '立即开通' : '购买套餐'}
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
      {channels.data && channels.data.length > 0 && (
        <div className='flex items-center gap-3 text-sm'>
          <span className='text-muted-foreground'>支付渠道</span>
          <Select
            value={channelID || String(channels.data[0].id)}
            onValueChange={setChannelID}
          >
            <SelectTrigger className='w-56'>
              <SelectValue placeholder='选择支付渠道' />
            </SelectTrigger>
            <SelectContent>
              {channels.data.map((channel) => (
                <SelectItem key={channel.id} value={String(channel.id)}>
                  {channel.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}
    </div>
  );
}
