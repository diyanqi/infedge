'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { Loader2, Wifi } from 'lucide-react';
import { z } from 'zod';

import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { CustomDnsAccountService } from '@/lib/services/custom';
import type {
  DnsAccountMutationPayload,
  DnsAccountProviderType,
} from '@/lib/services/openflare';
import {
  buildDnsAccountAuthorization,
  DNS_ACCOUNT_CREDENTIALS,
  DnsAccountService,
} from '@/lib/services/openflare';

import { getErrorMessage } from './website-utils';

const dnsAccountSchema = z
  .object({
    name: z
      .string()
      .trim()
      .min(1, '请输入名称')
      .max(255, '名称不能超过 255 个字符'),
    type: z.string().min(1),
    credentials: z.record(z.string(), z.string()),
  })
  .superRefine((value, context) => {
    const config =
      DNS_ACCOUNT_CREDENTIALS[value.type as DnsAccountProviderType];
    if (!config) {
      context.addIssue({
        code: z.ZodIssueCode.custom,
        path: ['type'],
        message: '请选择 DNS 服务商',
      });
      return;
    }
    for (const field of config.fields) {
      if (field.required && !value.credentials[field.key]?.trim()) {
        context.addIssue({
          code: z.ZodIssueCode.custom,
          path: ['credentials', field.key],
          message: `请填写${field.label}`,
        });
      }
    }
  });

type DnsAccountFormValues = z.infer<typeof dnsAccountSchema>;

interface DnsAccountCreateDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated?: () => void;
  mode?: 'admin' | 'user';
}

export function DnsAccountCreateDialog({
  open,
  onOpenChange,
  onCreated,
  mode = 'admin',
}: DnsAccountCreateDialogProps) {
  const queryClient = useQueryClient();
  const [error, setError] = useState('');
  const [testError, setTestError] = useState('');
  const [testSuccess, setTestSuccess] = useState('');
  const dnsAccountsQueryKey =
    mode === 'admin'
      ? ['openflare', 'dns-accounts']
      : ['custom', 'dns-accounts'];
  const createService =
    mode === 'admin' ? DnsAccountService : CustomDnsAccountService;
  const form = useForm<DnsAccountFormValues>({
    resolver: zodResolver(dnsAccountSchema),
    defaultValues: {
      name: '',
      type: 'cloudflare',
      credentials: { api_token: '' },
    },
  });
  const watchedType = form.watch('type') as DnsAccountProviderType;
  const credentialConfig = DNS_ACCOUNT_CREDENTIALS[watchedType];

  const createMutation = useMutation({
    mutationFn: (payload: DnsAccountMutationPayload) =>
      createService.create(payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: dnsAccountsQueryKey });
      form.reset();
      setError('');
      setTestError('');
      setTestSuccess('');
      onCreated?.();
      onOpenChange(false);
    },
    onError: (err) => setError(getErrorMessage(err)),
  });

  const testMutation = useMutation({
    mutationFn: (payload: DnsAccountMutationPayload) =>
      createService.test(payload),
    onSuccess: () => {
      setTestSuccess('连接成功，凭据有效');
      setTestError('');
    },
    onError: (err) => {
      setTestError(getErrorMessage(err));
      setTestSuccess('');
    },
  });

  const buildPayload = (
    values: DnsAccountFormValues,
  ): DnsAccountMutationPayload => ({
    name: values.name.trim(),
    type: values.type,
    authorization: buildDnsAccountAuthorization(
      values.type,
      values.credentials,
    ),
  });

  const onSubmit = form.handleSubmit((values) => {
    setError('');
    createMutation.mutate(buildPayload(values));
  });

  const handleTest = form.handleSubmit((values) => {
    setTestError('');
    setTestSuccess('');
    testMutation.mutate(buildPayload(values));
  });

  const handleProviderChange = (value: string) => {
    form.setValue('type', value);
    form.setValue('credentials', {});
  };

  const handleClose = () => {
    form.reset();
    setError('');
    setTestError('');
    setTestSuccess('');
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={(next) => !next && handleClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>添加 DNS 账号</DialogTitle>
          <DialogDescription>
            统一管理 DNS 服务商账号，用于 ACME 证书的 DNS 验证申请。
          </DialogDescription>
        </DialogHeader>

        <form className='space-y-4' onSubmit={onSubmit}>
          {error ? <p className='text-sm text-destructive'>{error}</p> : null}
          {testError ? (
            <p className='text-sm text-destructive'>{testError}</p>
          ) : null}
          {testSuccess ? (
            <p className='text-sm text-emerald-600'>{testSuccess}</p>
          ) : null}

          <div className='space-y-2'>
            <Label>账号名称</Label>
            <Input
              placeholder='Cloudflare 邮箱账号'
              {...form.register('name')}
            />
            {form.formState.errors.name ? (
              <p className='text-xs text-destructive'>
                {form.formState.errors.name.message}
              </p>
            ) : null}
          </div>

          <div className='space-y-2'>
            <Label>DNS 服务商</Label>
            <Select
              value={form.watch('type')}
              onValueChange={handleProviderChange}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {Object.values(DNS_ACCOUNT_CREDENTIALS).map((config) => (
                  <SelectItem key={config.provider} value={config.provider}>
                    {config.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {form.formState.errors.type ? (
              <p className='text-xs text-destructive'>
                {form.formState.errors.type.message}
              </p>
            ) : null}
          </div>

          <div className='space-y-3'>
            {credentialConfig?.fields.map((field) => (
              <div className='space-y-2' key={field.key}>
                <Label>{field.label}</Label>
                <Input
                  type={field.type ?? 'text'}
                  placeholder={field.placeholder}
                  {...form.register(`credentials.${field.key}`)}
                />
                {field.help ? (
                  <p className='text-xs text-muted-foreground'>{field.help}</p>
                ) : null}
                {form.formState.errors.credentials?.[field.key]?.message ? (
                  <p className='text-xs text-destructive'>
                    {form.formState.errors.credentials[field.key]?.message}
                  </p>
                ) : null}
              </div>
            ))}
          </div>

          <DialogFooter>
            <Button type='button' variant='outline' onClick={handleClose}>
              取消
            </Button>
            <Button
              type='button'
              variant='outline'
              onClick={() => void handleTest()}
              disabled={testMutation.isPending || createMutation.isPending}
            >
              {testMutation.isPending ? (
                <>
                  <Loader2 className='mr-1 size-3.5 animate-spin' />
                  测试中...
                </>
              ) : (
                <>
                  <Wifi className='mr-1 size-3.5' />
                  测试连接
                </>
              )}
            </Button>
            <Button
              type='submit'
              disabled={createMutation.isPending || testMutation.isPending}
            >
              {createMutation.isPending ? (
                <>
                  <Loader2 className='mr-1 size-3.5 animate-spin' />
                  提交中...
                </>
              ) : (
                '提交'
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
