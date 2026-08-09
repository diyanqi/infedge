// 站点控制台：概览 Tab。
'use client';

import { useState } from 'react';
import {
  CheckCircle2,
  Globe,
  Rocket,
  Server,
  ShieldCheck,
  Trash2,
} from 'lucide-react';

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';

import type { useSiteConsole } from '../use-site-console';

export function OverviewTab({
  console: value,
}: {
  console: ReturnType<typeof useSiteConsole>;
}) {
  const {
    route,
    selectedDomain,
    selectedZone,
    domainVerified,
    publish,
    deleteZone,
  } = value;
  const [confirmDelete, setConfirmDelete] = useState(false);

  const domain = selectedDomain?.domain ?? route?.site_name ?? '未命名站点';
  const running = Boolean(
    route?.enabled && route?.zone_domains?.length && domainVerified,
  );
  const status = running ? '运行中' : route ? '待配置' : '待验证';

  return (
    <div className='space-y-5'>
      <Card>
        <CardHeader>
          <div className='flex flex-wrap items-start justify-between gap-3'>
            <div className='flex flex-col gap-1'>
              <CardTitle className='flex items-center gap-2 text-base'>
                <Globe className='size-4 text-primary' />
                {domain}
              </CardTitle>
              <CardDescription>
                {selectedZone
                  ? `域名分组：${selectedZone.domain}`
                  : '尚未关联域名分组'}
              </CardDescription>
            </div>
            <Badge variant={running ? 'secondary' : 'outline'}>{status}</Badge>
          </div>
        </CardHeader>
        <CardContent className='grid gap-4 sm:grid-cols-3'>
          <Info
            icon={Server}
            label='源站'
            value={
              route?.origin_url ||
              (route?.origin_id ? `源站 #${route.origin_id}` : '未配置')
            }
          />
          <Info
            icon={ShieldCheck}
            label='HTTPS'
            value={route?.enable_https ? '已启用' : '未启用'}
          />
          <Info
            icon={CheckCircle2}
            label='域名验证'
            value={domainVerified ? '已验证' : '待验证'}
          />
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className='text-base'>发布与删除</CardTitle>
          <CardDescription>
            配置保存后需要部署才会生效；删除站点会同时删除该域名分组下的全部域名与站点。
          </CardDescription>
        </CardHeader>
        <CardContent className='flex flex-wrap items-center gap-3'>
          <Button onClick={() => publish.mutate()} disabled={publish.isPending}>
            <Rocket data-icon='inline-start' />
            {publish.isPending ? '部署中…' : '部署'}
          </Button>
          <Button variant='destructive' onClick={() => setConfirmDelete(true)}>
            <Trash2 data-icon='inline-start' />
            删除站点
          </Button>
        </CardContent>
      </Card>

      <AlertDialog open={confirmDelete} onOpenChange={setConfirmDelete}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除站点？</AlertDialogTitle>
            <AlertDialogDescription>
              即将删除 {domain} 及其所属域名分组（
              {selectedZone?.domain ?? '未知'}）。
              分组下的全部域名和站点配置都会被删除，此操作不可恢复。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction
              onClick={(event) => {
                event.preventDefault();
                deleteZone.mutate();
              }}
              className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
            >
              {deleteZone.isPending ? '删除中…' : '确认删除'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function Info({
  icon: Icon,
  label,
  value,
}: {
  icon: typeof Globe;
  label: string;
  value: string;
}) {
  return (
    <div className='rounded-md border bg-muted/20 p-3'>
      <p className='flex items-center gap-1.5 text-xs text-muted-foreground'>
        <Icon className='size-3.5' />
        {label}
      </p>
      <p className='mt-2 truncate text-sm font-medium' title={value}>
        {value || '未配置'}
      </p>
    </div>
  );
}
