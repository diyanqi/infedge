'use client';

import Link from 'next/link';
import { useMemo, useState } from 'react';
import { useQueries, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  ArrowRight,
  CheckCircle2,
  CircleAlert,
  Cloud,
  Globe,
  Plus,
  RefreshCw,
  Search,
  Server,
  Settings2,
} from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { LoadingStateWithBorder } from '@/components/layout/loading';
import { ErrorInline } from '@/components/layout/error';
import { EmptyStateWithBorder } from '@/components/layout/empty';
import { CustomService } from '@/lib/services/custom';
import type { ResourceRoute } from '@/lib/services/custom';

function routeStatus(route: ResourceRoute) {
  if (!route.enabled) return { label: '已停用', variant: 'outline' as const };
  if (!route.zone_domains.length)
    return { label: '待配置域名', variant: 'outline' as const };
  return { label: '运行中', variant: 'secondary' as const };
}

function primaryDomain(route: ResourceRoute) {
  return (
    route.zone_domains[0]?.domain || route.site_name || `站点 #${route.id}`
  );
}

function SiteRow({ route }: { route: ResourceRoute }) {
  const status = routeStatus(route);
  const domains = route.zone_domains.map((item) => item.domain).join('、');
  return (
    <div className='flex flex-col gap-4 border-b px-4 py-4 last:border-b-0 sm:flex-row sm:items-center sm:justify-between'>
      <div className='flex min-w-0 items-start gap-3'>
        <div className='flex size-9 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary'>
          <Cloud className='size-4' />
        </div>
        <div className='min-w-0'>
          <div className='flex flex-wrap items-center gap-2'>
            <p className='truncate text-sm font-medium'>
              {route.site_name || primaryDomain(route)}
            </p>
            <Badge variant={status.variant}>{status.label}</Badge>
          </div>
          <p className='mt-1 truncate text-xs text-muted-foreground'>
            {domains || '尚未绑定域名'}
          </p>
        </div>
      </div>
      <div className='flex items-center gap-2 sm:shrink-0'>
        <span className='hidden text-xs text-muted-foreground md:inline'>
          ID {route.id}
        </span>
        <Button variant='outline' size='sm' asChild>
          <Link href={`/resources/detail?id=${route.id}`}>
            <Settings2 data-icon='inline-start' />
            管理站点
          </Link>
        </Button>
      </div>
    </div>
  );
}

export default function ResourcesPage() {
  const queryClient = useQueryClient();
  const [keyword, setKeyword] = useState('');
  const routes = useQuery({
    queryKey: ['custom', 'routes'],
    queryFn: () => CustomService.listRoutes(),
  });
  const zones = useQuery({
    queryKey: ['custom', 'zones'],
    queryFn: () => CustomService.listZones(),
  });
  const origins = useQuery({
    queryKey: ['custom', 'origins'],
    queryFn: () => CustomService.listOrigins(),
  });
  const zoneDetails = useQueries({
    queries: (zones.data ?? []).map((zone) => ({
      queryKey: ['custom', 'zone', zone.id],
      queryFn: () => CustomService.getZone(zone.id),
    })),
  });
  const domainCount = zoneDetails.reduce(
    (count, query) => count + (query.data?.domains.length ?? 0),
    0,
  );
  const filteredRoutes = useMemo(() => {
    const value = keyword.trim().toLowerCase();
    if (!value) return routes.data ?? [];
    return (routes.data ?? []).filter((route) =>
      [
        route.site_name,
        ...route.zone_domains.map((domain) => domain.domain),
        route.origin_url,
      ]
        .join(' ')
        .toLowerCase()
        .includes(value),
    );
  }, [keyword, routes.data]);

  const refresh = () => {
    void queryClient.invalidateQueries({ queryKey: ['custom'] });
  };

  if (routes.isLoading || zones.isLoading || origins.isLoading) {
    return (
      <div className='w-full px-1 py-6'>
        <LoadingStateWithBorder
          title='加载站点'
          description='正在读取你的网站和接入配置...'
        />
      </div>
    );
  }

  if (routes.isError || zones.isError || origins.isError) {
    return (
      <div className='w-full space-y-6 px-1 py-6'>
        <PageTitle />
        <ErrorInline message='站点数据加载失败' onRetry={refresh} />
      </div>
    );
  }

  return (
    <div className='w-full space-y-6 px-1 py-6'>
      <div className='flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between'>
        <div>
          <PageTitle />
          <p className='mt-2 text-sm text-muted-foreground'>
            集中管理网站、域名和源站，进入站点后再调整加速与安全配置。
          </p>
        </div>
        <div className='flex items-center gap-2'>
          <Button variant='outline' size='sm' onClick={refresh}>
            <RefreshCw data-icon='inline-start' />
            刷新
          </Button>
          <Button size='sm' asChild>
            <Link href='/resources/configure?new=1'>
              <Plus data-icon='inline-start' />
              新增站点
            </Link>
          </Button>
        </div>
      </div>

      <section className='grid overflow-hidden rounded-lg border bg-card sm:grid-cols-3'>
        <Summary
          icon={Globe}
          label='网站'
          value={routes.data?.length ?? 0}
          detail='已创建的加速站点'
        />
        <Summary
          icon={CheckCircle2}
          label='运行中'
          value={
            (routes.data ?? []).filter(
              (route) => route.enabled && route.zone_domains.length > 0,
            ).length
          }
          detail='已绑定域名并启用'
        />
        <Summary
          icon={Server}
          label='源站'
          value={origins.data?.length ?? 0}
          detail={`${domainCount} 个托管域名`}
        />
      </section>

      <section className='overflow-hidden rounded-lg border bg-card'>
        <div className='flex flex-col gap-3 border-b px-4 py-4 sm:flex-row sm:items-center sm:justify-between'>
          <div>
            <h2 className='text-base font-semibold'>我的网站</h2>
            <p className='mt-1 text-xs text-muted-foreground'>
              选择一个网站进入站点控制台。
            </p>
          </div>
          <div className='relative w-full sm:w-64'>
            <Search className='absolute left-2.5 top-2.5 size-4 text-muted-foreground' />
            <Input
              value={keyword}
              onChange={(event) => setKeyword(event.target.value)}
              placeholder='搜索网站或域名'
              className='pl-8'
            />
          </div>
        </div>
        {filteredRoutes.length ? (
          filteredRoutes.map((route) => (
            <SiteRow key={route.id} route={route} />
          ))
        ) : (
          <div className='p-4'>
            <EmptyStateWithBorder
              icon={Globe}
              title={keyword ? '没有匹配的网站' : '还没有网站'}
              description={
                keyword
                  ? '试试其他网站名称或域名。'
                  : '添加一个网站，开始配置域名、源站和安全加速。'
              }
            />
          </div>
        )}
      </section>

      <section className='grid gap-4 md:grid-cols-2'>
        <Link
          href='/resources/configure?new=1'
          className='group flex items-center gap-3 rounded-lg border bg-card p-4 transition-colors hover:bg-muted/30'
        >
          <div className='flex size-9 items-center justify-center rounded-md bg-primary/10 text-primary'>
            <Plus className='size-4' />
          </div>
          <div className='min-w-0 flex-1'>
            <p className='text-sm font-medium'>接入新网站</p>
            <p className='mt-1 text-xs text-muted-foreground'>
              添加根域、子域和源站
            </p>
          </div>
          <ArrowRight className='size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5' />
        </Link>
        <div className='flex items-center gap-3 rounded-lg border bg-card p-4'>
          <div className='flex size-9 items-center justify-center rounded-md bg-muted text-muted-foreground'>
            <CircleAlert className='size-4' />
          </div>
          <div>
            <p className='text-sm font-medium'>接入提示</p>
            <p className='mt-1 text-xs text-muted-foreground'>
              域名需将 CNAME 指向平台提供的接入地址。
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}

function PageTitle() {
  return (
    <div className='flex items-center gap-2'>
      <Globe className='size-5 text-primary' />
      <h1 className='text-2xl font-semibold tracking-tight'>网站安全加速</h1>
    </div>
  );
}

function Summary({
  icon: Icon,
  label,
  value,
  detail,
}: {
  icon: typeof Globe;
  label: string;
  value: number;
  detail: string;
}) {
  return (
    <div className='flex items-center gap-3 border-b px-4 py-4 last:border-b-0 sm:border-b-0 sm:border-r sm:last:border-r-0'>
      <Icon className='size-5 text-primary' />
      <div>
        <p className='text-xs text-muted-foreground'>{label}</p>
        <p className='mt-1 text-xl font-semibold tabular-nums'>{value}</p>
        <p className='text-[11px] text-muted-foreground'>{detail}</p>
      </div>
    </div>
  );
}
