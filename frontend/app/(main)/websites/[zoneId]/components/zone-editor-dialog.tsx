'use client';

import { useEffect } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { Loader2 } from 'lucide-react';
import { toast } from 'sonner';
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
  SiteService,
  ZoneService,
  zoneQueryKey,
  type SiteItem,
  type ZoneItem,
} from '@/lib/services/openflare';

const schema = z.object({
  domain: z
    .string()
    .trim()
    .min(1, '请输入域名')
    .refine(
      (value) => !/[*/?#@]|:\/\//.test(value),
      '请输入不含协议或通配符的根域或子域',
    ),
});
type Values = z.infer<typeof schema>;

export function ZoneEditorDialog({
  open,
  onOpenChange,
  zone,
  onCreated,
}: {
  open: boolean;
  onOpenChange(open: boolean): void;
  zone?: ZoneItem | null;
  onCreated?(site: SiteItem): void;
}) {
  const queryClient = useQueryClient();
  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: { domain: '' },
  });
  useEffect(() => {
    if (open) form.reset({ domain: zone?.domain ?? '' });
  }, [form, open, zone]);
  const mutation = useMutation<ZoneItem | SiteItem, Error, Values>({
    mutationFn: async (values) => {
      if (zone) {
        return ZoneService.update(zone.id, {
          domain: values.domain.toLowerCase(),
        });
      }
      return SiteService.create({
        domain: values.domain.toLowerCase(),
      });
    },
    onSuccess: async (result) => {
      const site = !zone && 'zone' in result ? result : undefined;
      toast.success(zone ? 'Zone 已更新' : '域名已添加，请完成 TXT 验证');
      await queryClient.invalidateQueries({ queryKey: zoneQueryKey });
      if (site) onCreated?.(site);
      onOpenChange(false);
    },
    onError: (error) =>
      toast.error(error instanceof Error ? error.message : '保存失败'),
  });
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{zone ? '编辑 Zone' : '新增 CDN 域名'}</DialogTitle>
          <DialogDescription>
            直接输入根域或子域，例如 example.com 或 www.example.com。
          </DialogDescription>
        </DialogHeader>
        <form
          id='zone-editor'
          className='space-y-4'
          onSubmit={form.handleSubmit((values) =>
            mutation.mutate({ domain: values.domain.toLowerCase() }),
          )}
        >
          <div className='space-y-1.5'>
            <Label htmlFor='zone-domain'>域名</Label>
            <Input
              id='zone-domain'
              placeholder='example.com 或 www.example.com'
              {...form.register('domain')}
            />
            {form.formState.errors.domain && (
              <p className='text-xs text-destructive'>
                {form.formState.errors.domain.message}
              </p>
            )}
          </div>
        </form>
        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button
            form='zone-editor'
            type='submit'
            disabled={mutation.isPending}
          >
            {mutation.isPending && (
              <Loader2 className='mr-1 size-4 animate-spin' />
            )}
            {zone ? '保存修改' : '添加域名'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
