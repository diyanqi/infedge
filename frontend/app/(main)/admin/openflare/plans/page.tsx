'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Copy,
  CreditCard,
  Package,
  Pencil,
  Plus,
  Ticket,
  Trash2,
} from 'lucide-react';
import { toast } from 'sonner';
import { CustomService } from '@/lib/services/custom';
import type { ChannelInput, PlanInput } from '@/lib/services/custom';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

const initialPlan: PlanInput = {
  name: '',
  description: '',
  price_fen: 0,
  billing_months: 1,
  high_speed_bytes: 10737418240,
  throttle_bytes_per_sec: 1048576,
  daily_publish_limit: 1,
  max_zones: 1,
  max_origins: 1,
  max_routes: 1,
  max_pages: 0,
  enabled: true,
};
const initialChannel: ChannelInput = {
  name: '',
  gateway: '',
  pid: '',
  secret_key: '',
  enabled: true,
  sort: 0,
};

const formatBytes = (value: number) =>
  value <= 0 ? '不限' : `${(value / 1073741824).toFixed(2)} GB / 月`;

export default function AdminPlansPage() {
  const qc = useQueryClient();
  const [plan, setPlan] = useState(initialPlan);
  const [editingPlan, setEditingPlan] = useState<number | null>(null);
  const [channel, setChannel] = useState(initialChannel);
  const [editingChannel, setEditingChannel] = useState<number | null>(null);
  const [redeemPlanID, setRedeemPlanID] = useState('');
  const plans = useQuery({
    queryKey: ['custom', 'admin-plans'],
    queryFn: () => CustomService.adminListPlans(),
  });
  const channels = useQuery({
    queryKey: ['custom', 'admin-channels'],
    queryFn: () => CustomService.adminListChannels(),
  });
  const redeemCodes = useQuery({
    queryKey: ['custom', 'admin-redeem-codes'],
    queryFn: () => CustomService.adminListRedeemCodes(),
  });
  const savePlan = useMutation({
    mutationFn: () =>
      editingPlan
        ? CustomService.adminUpdatePlan(editingPlan, plan)
        : CustomService.adminCreatePlan(plan),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ['custom', 'admin-plans'] });
      setPlan(initialPlan);
      setEditingPlan(null);
      toast.success('套餐已保存');
    },
    onError: (e) => toast.error(e instanceof Error ? e.message : '保存失败'),
  });
  const saveChannel = useMutation({
    mutationFn: () =>
      editingChannel
        ? CustomService.adminUpdateChannel(editingChannel, channel)
        : CustomService.adminCreateChannel(channel),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ['custom', 'admin-channels'] });
      setChannel(initialChannel);
      setEditingChannel(null);
      toast.success('支付渠道已保存');
    },
    onError: (e) => toast.error(e instanceof Error ? e.message : '保存失败'),
  });
  const removePlan = useMutation({
    mutationFn: (id: number) => CustomService.adminDeletePlan(id),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ['custom', 'admin-plans'] }),
  });
  const removeChannel = useMutation({
    mutationFn: (id: number) => CustomService.adminDeleteChannel(id),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ['custom', 'admin-channels'] }),
  });
  const createRedeemCode = useMutation({
    mutationFn: () => CustomService.adminCreateRedeemCode(Number(redeemPlanID)),
    onSuccess: async (code) => {
      await redeemCodes.refetch();
      toast.success(`兑换码已创建：${code.code}`);
    },
    onError: (e) =>
      toast.error(e instanceof Error ? e.message : '创建兑换码失败'),
  });
  const numberField = (label: string, key: keyof PlanInput) => (
    <div>
      <Label>{label}</Label>
      <Input
        type='number'
        min='0'
        value={String(plan[key])}
        onChange={(e) => setPlan({ ...plan, [key]: Number(e.target.value) })}
      />
    </div>
  );
  const throttleKB = Math.max(
    0,
    Math.round(plan.throttle_bytes_per_sec / 1024),
  );

  return (
    <div className='w-full space-y-6 py-6 px-1'>
      <div className='flex items-center gap-2'>
        <Package className='size-5 text-primary' />
        <h1 className='text-2xl font-semibold tracking-tight'>套餐与支付</h1>
      </div>
      <Card>
        <CardHeader>
          <CardTitle className='flex items-center gap-2 text-base'>
            <Ticket className='size-4' />
            创建兑换码
          </CardTitle>
        </CardHeader>
        <CardContent className='flex flex-col gap-3 sm:flex-row sm:items-end'>
          <div className='w-full sm:max-w-sm'>
            <Label>兑换套餐</Label>
            <Select value={redeemPlanID} onValueChange={setRedeemPlanID}>
              <SelectTrigger>
                <SelectValue placeholder='选择启用中的套餐' />
              </SelectTrigger>
              <SelectContent>
                {(plans.data ?? [])
                  .filter((item) => item.enabled)
                  .map((item) => (
                    <SelectItem key={item.id} value={String(item.id)}>
                      {item.name}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
          </div>
          <Button
            onClick={() => createRedeemCode.mutate()}
            disabled={!redeemPlanID || createRedeemCode.isPending}
          >
            <Plus className='mr-2 size-4' />
            {createRedeemCode.isPending ? '生成中...' : '生成兑换码'}
          </Button>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle className='text-base'>兑换码记录</CardTitle>
        </CardHeader>
        <CardContent className='space-y-2'>
          {(redeemCodes.data ?? []).map((item) => (
            <div
              key={item.id}
              className='flex flex-wrap items-center justify-between gap-2 rounded-md border p-3 text-sm'
            >
              <div>
                <p className='font-mono font-medium'>{item.code}</p>
                <p className='text-xs text-muted-foreground'>
                  {item.plan?.name ?? `套餐 #${item.plan_id}`} ·{' '}
                  {item.status === 'used'
                    ? `已使用${item.used_at ? ` · ${new Date(item.used_at).toLocaleString()}` : ''}`
                    : '未使用'}
                </p>
              </div>
              <Button
                variant='ghost'
                size='icon'
                title='复制兑换码'
                aria-label='复制兑换码'
                disabled={item.status === 'used'}
                onClick={() => {
                  void navigator.clipboard.writeText(item.code);
                  toast.success('兑换码已复制');
                }}
              >
                <Copy className='size-4' />
              </Button>
            </div>
          ))}
          {!redeemCodes.isLoading && !(redeemCodes.data ?? []).length ? (
            <p className='text-sm text-muted-foreground'>暂无兑换码</p>
          ) : null}
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle className='text-base'>
            {editingPlan ? '编辑套餐' : '新增套餐'}
          </CardTitle>
        </CardHeader>
        <CardContent className='grid gap-3 md:grid-cols-3'>
          <div>
            <Label>套餐名称</Label>
            <Input
              value={plan.name}
              onChange={(e) => setPlan({ ...plan, name: e.target.value })}
            />
          </div>
          <div>
            <Label>说明</Label>
            <Input
              value={plan.description}
              onChange={(e) =>
                setPlan({ ...plan, description: e.target.value })
              }
            />
          </div>
          {numberField('价格（分）', 'price_fen')}
          {numberField('周期（月）', 'billing_months')}
          {numberField('高速流量（字节）', 'high_speed_bytes')}
          <div>
            <Label>超额速度（KB/s）</Label>
            <Input
              type='number'
              min='0'
              value={String(throttleKB)}
              onChange={(e) =>
                setPlan({
                  ...plan,
                  throttle_bytes_per_sec:
                    Math.max(0, Number(e.target.value)) * 1024,
                })
              }
            />
          </div>
          {numberField('每日发布次数，0不限', 'daily_publish_limit')}
          {numberField('域名分组数量', 'max_zones')}
          {numberField('源站数量', 'max_origins')}
          {numberField('规则数量', 'max_routes')}
          {numberField('Pages 数量', 'max_pages')}
          <div className='flex items-center gap-2 pt-6'>
            <Switch
              checked={plan.enabled}
              onCheckedChange={(enabled) => setPlan({ ...plan, enabled })}
            />
            <Label>启用</Label>
          </div>
          <div className='flex items-end gap-2'>
            <Button
              onClick={() => savePlan.mutate()}
              disabled={!plan.name || savePlan.isPending}
            >
              <Plus className='mr-2 size-4' />
              保存套餐
            </Button>
            {editingPlan && (
              <Button
                variant='outline'
                onClick={() => {
                  setPlan(initialPlan);
                  setEditingPlan(null);
                }}
              >
                取消
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
      <div className='grid gap-3 md:grid-cols-2 xl:grid-cols-3'>
        {(plans.data ?? []).map((item) => (
          <Card key={item.id} className='border-dashed'>
            <CardContent className='space-y-2 pt-6'>
              <div className='flex items-center justify-between'>
                <p className='font-medium'>{item.name}</p>
                <span className='text-sm'>
                  {item.price_fen === 0
                    ? '免费'
                    : '¥' + (item.price_fen / 100).toFixed(2)}
                </span>
              </div>
              <p className='text-xs text-muted-foreground'>
                高速 {formatBytes(item.high_speed_bytes)} · 超额{' '}
                {item.throttle_bytes_per_sec <= 0
                  ? '不额外限速'
                  : `${(item.throttle_bytes_per_sec / 1024).toFixed(1)} KB/s`}
              </p>
              <div className='flex gap-2'>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={() => {
                    setPlan(item);
                    setEditingPlan(item.id);
                  }}
                >
                  <Pencil className='mr-1 size-3' />
                  编辑
                </Button>
                <Button
                  variant='ghost'
                  size='icon'
                  onClick={() => removePlan.mutate(item.id)}
                  aria-label='删除套餐'
                >
                  <Trash2 className='size-4' />
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
      <Card>
        <CardHeader>
          <CardTitle className='flex items-center gap-2 text-base'>
            <CreditCard className='size-4' />
            易支付渠道
          </CardTitle>
        </CardHeader>
        <CardContent className='grid gap-3 md:grid-cols-3'>
          <div>
            <Label>名称</Label>
            <Input
              value={channel.name}
              onChange={(e) => setChannel({ ...channel, name: e.target.value })}
            />
          </div>
          <div>
            <Label>网关地址</Label>
            <Input
              value={channel.gateway}
              onChange={(e) =>
                setChannel({ ...channel, gateway: e.target.value })
              }
              placeholder='https://pay.example.com'
            />
          </div>
          <div>
            <Label>商户 PID</Label>
            <Input
              value={channel.pid}
              onChange={(e) => setChannel({ ...channel, pid: e.target.value })}
            />
          </div>
          <div>
            <Label>密钥</Label>
            <Input
              type='password'
              value={channel.secret_key}
              onChange={(e) =>
                setChannel({ ...channel, secret_key: e.target.value })
              }
              placeholder={editingChannel ? '留空保留旧密钥' : ''}
            />
          </div>
          <div>
            <Label>排序</Label>
            <Input
              type='number'
              value={String(channel.sort)}
              onChange={(e) =>
                setChannel({ ...channel, sort: Number(e.target.value) })
              }
            />
          </div>
          <div className='flex items-center gap-2 pt-6'>
            <Switch
              checked={channel.enabled}
              onCheckedChange={(enabled) => setChannel({ ...channel, enabled })}
            />
            <Label>启用</Label>
          </div>
          <div className='flex items-end gap-2'>
            <Button
              onClick={() => saveChannel.mutate()}
              disabled={!channel.name || !channel.gateway || !channel.pid}
            >
              <Plus className='mr-2 size-4' />
              保存渠道
            </Button>
            {editingChannel && (
              <Button
                variant='outline'
                onClick={() => {
                  setChannel(initialChannel);
                  setEditingChannel(null);
                }}
              >
                取消
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
      <div className='grid gap-3 md:grid-cols-2'>
        {(channels.data ?? []).map((item) => (
          <Card key={item.id} className='border-dashed'>
            <CardContent className='flex items-center justify-between gap-3 pt-6'>
              <div>
                <p className='font-medium'>{item.name}</p>
                <p className='text-sm text-muted-foreground'>
                  {item.gateway} · PID {item.pid}
                </p>
              </div>
              <div className='flex gap-2'>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={() => {
                    setChannel({ ...item, secret_key: '' });
                    setEditingChannel(item.id);
                  }}
                >
                  <Pencil className='mr-1 size-3' />
                  编辑
                </Button>
                <Button
                  variant='ghost'
                  size='icon'
                  onClick={() => removeChannel.mutate(item.id)}
                  aria-label='删除渠道'
                >
                  <Trash2 className='size-4' />
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
