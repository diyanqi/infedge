'use client';

import Link from 'next/link';
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  ArrowRight,
  BarChart3,
  CheckCircle2,
  Cloud,
  FileText,
  Globe,
  Layers3,
  Plus,
  RefreshCw,
  Server,
} from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { EmptyStateWithBorder } from '@/components/layout/empty';
import { ErrorInline } from '@/components/layout/error';
import { LoadingStateWithBorder } from '@/components/layout/loading';
import { CustomService } from '@/lib/services/custom';
import type { ResourceRoute } from '@/lib/services/custom';
import { PagesService } from '@/lib/services/openflare';

function status(route: ResourceRoute) {
  if (!route.enabled) return { label: '已停用', variant: 'outline' as const };
  if (!route.zone_domains.length)
    return { label: '待配置域名', variant: 'outline' as const };
  return { label: '运行中', variant: 'secondary' as const };
}

export function OrdinaryDashboard() {
  const [tab, setTab] = useState('acceleration');
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
  const pages = useQuery({
    queryKey: ['openflare', 'pages', 'projects'],
    queryFn: () => PagesService.listProjects(),
  });
  const busy =
    routes.isLoading || zones.isLoading || origins.isLoading || pages.isLoading;
  const error = routes.error || zones.error || origins.error || pages.error;
  const refresh = () => {
    void routes.refetch();
    void zones.refetch();
    void origins.refetch();
    void pages.refetch();
  };

  if (busy)
    return (
      <div className='w-full px-1 py-6'>
        <LoadingStateWithBorder
          title='加载服务总览'
          description='正在读取网站和 Pages 项目...'
        />
      </div>
    );
  if (error)
    return (
      <div className='w-full space-y-6 px-1 py-6'>
        <PageTitle />
        <ErrorInline
          message={error instanceof Error ? error.message : '服务数据加载失败'}
          onRetry={refresh}
        />
      </div>
    );

  return (
    <div className='w-full space-y-6 px-1 py-6'>
      <div className='flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between'>
        <div>
          <PageTitle />
          <p className='mt-2 text-sm text-muted-foreground'>
            从这里开始管理网站安全加速和静态网站。
          </p>
        </div>
        <Button variant='outline' size='sm' onClick={refresh}>
          <RefreshCw data-icon='inline-start' />
          刷新
        </Button>
      </div>
      <Tabs value={tab} onValueChange={setTab} className='w-full'>
        <TabsList
          variant='line'
          className='w-full justify-start gap-5 border-b rounded-none px-0'
        >
          <TabsTrigger value='acceleration' className='flex-none px-1'>
            网站安全加速
          </TabsTrigger>
          <TabsTrigger value='makers' className='flex-none px-1'>
            Pages
          </TabsTrigger>
        </TabsList>
        <TabsContent value='acceleration' className='mt-5 space-y-5'>
          <AccelerationOverview
            routes={routes.data ?? []}
            zones={zones.data ?? []}
            origins={origins.data ?? []}
          />
        </TabsContent>
        <TabsContent value='makers' className='mt-5 space-y-5'>
          <PagesOverview count={pages.data?.length ?? 0} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function AccelerationOverview({
  routes,
  zones,
  origins,
}: {
  routes: ResourceRoute[];
  zones: { id: number }[];
  origins: { id: number }[];
}) {
  const running = routes.filter(
    (route) => route.enabled && route.zone_domains.length > 0,
  ).length;
  const domains = routes.reduce(
    (count, route) => count + route.zone_domains.length,
    0,
  );
  return (
    <>
      <div className='grid overflow-hidden rounded-lg border bg-card sm:grid-cols-2 lg:grid-cols-4'>
        <Metric
          icon={Globe}
          label='网站'
          value={routes.length}
          hint='已创建站点'
        />
        <Metric
          icon={CheckCircle2}
          label='运行中'
          value={running}
          hint='域名已接入'
        />
        <Metric
          icon={Server}
          label='托管域名'
          value={domains}
          hint={`${zones.length} 个域名分组`}
        />
        <Metric
          icon={Cloud}
          label='源站'
          value={origins.length}
          hint='可复用源站地址'
        />
      </div>
      <div className='grid gap-5 lg:grid-cols-[minmax(0,1.5fr)_minmax(260px,0.75fr)]'>
        <section className='overflow-hidden rounded-lg border bg-card'>
          <div className='flex items-center justify-between border-b px-4 py-4'>
            <div>
              <h2 className='text-base font-semibold'>网站列表</h2>
              <p className='mt-1 text-xs text-muted-foreground'>
                选择网站进入站点控制台。
              </p>
            </div>
            <Button size='sm' asChild>
              <Link href='/resources/configure?new=1'>
                <Plus data-icon='inline-start' />
                新增网站
              </Link>
            </Button>
          </div>
          {routes.length ? (
            <div className='divide-y'>
              {routes.slice(0, 6).map((route) => (
                <SiteItem key={route.id} route={route} />
              ))}
            </div>
          ) : (
            <div className='p-4'>
              <EmptyStateWithBorder
                icon={Globe}
                title='还没有网站'
                description='添加网站后，在这里查看接入状态和配置入口。'
              />
            </div>
          )}
          {routes.length > 6 && (
            <div className='border-t px-4 py-3'>
              <Button variant='ghost' size='sm' asChild>
                <Link href='/resources'>
                  查看全部网站
                  <ArrowRight data-icon='inline-end' />
                </Link>
              </Button>
            </div>
          )}
        </section>
        <SetupPanel
          hasDomainGroup={zones.length > 0}
          hasOrigin={origins.length > 0}
          hasRoute={running > 0}
        />
      </div>
    </>
  );
}

function PagesOverview({ count }: { count: number }) {
  const projects = useQuery({
    queryKey: ['openflare', 'pages', 'projects'],
    queryFn: () => PagesService.listProjects(),
  });
  return (
    <>
      <div className='grid overflow-hidden rounded-lg border bg-card sm:grid-cols-3'>
        <Metric
          icon={FileText}
          label='Pages 项目'
          value={count}
          hint='静态网站项目'
        />
        <Metric
          icon={BarChart3}
          label='近 7 天访问'
          value={0}
          hint='暂无统计数据'
        />
        <Metric
          icon={CheckCircle2}
          label='已部署项目'
          value={
            (projects.data ?? []).filter((item) =>
              Boolean(item.active_deployment_id),
            ).length
          }
          hint='当前有线上版本'
        />
      </div>
      <section className='overflow-hidden rounded-lg border bg-card'>
        <div className='flex items-center justify-between border-b px-4 py-4'>
          <div>
            <h2 className='text-base font-semibold'>最近的 Pages</h2>
            <p className='mt-1 text-xs text-muted-foreground'>
              Pages 用于托管和部署静态网站项目。
            </p>
          </div>
          <Button size='sm' asChild>
            <Link href='/pages'>
              <Plus data-icon='inline-start' />
              创建项目
            </Link>
          </Button>
        </div>
        {count ? (
          <div className='divide-y'>
            {(projects.data ?? []).slice(0, 6).map((project) => (
              <Link
                key={project.id}
                href={`/pages/detail?id=${project.id}`}
                className='flex items-center gap-3 px-4 py-4 hover:bg-muted/30'
              >
                <div className='flex size-9 items-center justify-center rounded-md bg-primary/10 text-primary'>
                  <FileText className='size-4' />
                </div>
                <div className='min-w-0 flex-1'>
                  <p className='truncate text-sm font-medium'>{project.name}</p>
                  <p className='mt-1 text-xs text-muted-foreground'>
                    {project.slug || '尚未设置访问标识'}
                  </p>
                </div>
                <ArrowRight className='size-4 text-muted-foreground' />
              </Link>
            ))}
          </div>
        ) : (
          <div className='p-4'>
            <EmptyStateWithBorder
              icon={FileText}
              title='还没有 Pages 项目'
              description='创建一个项目并上传静态网站，就能获得可访问的边缘站点。'
            />
          </div>
        )}
      </section>
    </>
  );
}

function SiteItem({ route }: { route: ResourceRoute }) {
  const state = status(route);
  const domain =
    route.zone_domains[0]?.domain || route.site_name || `站点 #${route.id}`;
  return (
    <div className='flex flex-col gap-3 px-4 py-4 sm:flex-row sm:items-center sm:justify-between'>
      <div className='flex min-w-0 items-start gap-3'>
        <div className='flex size-9 shrink-0 items-center justify-center rounded-md bg-muted'>
          <Globe className='size-4 text-primary' />
        </div>
        <div className='min-w-0'>
          <div className='flex flex-wrap items-center gap-2'>
            <p className='truncate text-sm font-medium'>
              {route.site_name || domain}
            </p>
            <Badge variant={state.variant}>{state.label}</Badge>
          </div>
          <p className='mt-1 truncate text-xs text-muted-foreground'>
            {route.zone_domains.map((item) => item.domain).join('、') ||
              '尚未绑定域名'}
          </p>
        </div>
      </div>
      <Button variant='outline' size='sm' asChild>
        <Link href={`/resources/detail?id=${route.id}`}>
          管理
          <ArrowRight data-icon='inline-end' />
        </Link>
      </Button>
    </div>
  );
}

function SetupPanel({
  hasDomainGroup,
  hasOrigin,
  hasRoute,
}: {
  hasDomainGroup: boolean;
  hasOrigin: boolean;
  hasRoute: boolean;
}) {
  const done = [hasDomainGroup, hasOrigin, hasRoute];
  const progress = Math.round(
    (done.filter(Boolean).length / done.length) * 100,
  );
  return (
    <section className='rounded-lg border bg-card p-4 sm:p-5'>
      <div className='flex items-start justify-between gap-3'>
        <div>
          <h2 className='text-base font-semibold'>接入进度</h2>
          <p className='mt-1 text-xs text-muted-foreground'>
            完成以下步骤即可开始加速。
          </p>
        </div>
        <span className='text-sm font-semibold tabular-nums'>{progress}%</span>
      </div>
      <div className='mt-4 h-1.5 overflow-hidden rounded-full bg-muted'>
        <div
          className='h-full rounded-full bg-primary'
          style={{ width: `${progress}%` }}
        />
      </div>
      <div className='mt-5 space-y-4'>
        <Step
          done={hasDomainGroup}
          title='添加域名'
          href='/resources/configure?new=1'
        />
        <Step
          done={hasOrigin}
          title='配置源站'
          href='/resources/configure?new=1'
        />
        <Step
          done={hasRoute}
          title='创建并接入网站'
          href='/resources/configure?new=1'
        />
      </div>
    </section>
  );
}

function Step({
  done,
  title,
  href,
}: {
  done: boolean;
  title: string;
  href: string;
}) {
  return (
    <Link
      href={href}
      className='flex items-center gap-3 rounded-md text-sm hover:bg-muted/40'
    >
      <span className='flex size-5 items-center justify-center'>
        {done ? (
          <CheckCircle2 className='size-4 text-emerald-600' />
        ) : (
          <span className='size-2 rounded-full bg-muted-foreground/40' />
        )}
      </span>
      <span
        className={done ? 'text-muted-foreground line-through' : 'font-medium'}
      >
        {title}
      </span>
      <ArrowRight className='ml-auto size-3.5 text-muted-foreground' />
    </Link>
  );
}
function Metric({
  icon: Icon,
  label,
  value,
  hint,
}: {
  icon: typeof Globe;
  label: string;
  value: number;
  hint: string;
}) {
  return (
    <div className='flex min-w-0 items-center gap-3 border-b px-4 py-4 last:border-b-0 sm:border-b-0 sm:border-r sm:last:border-r-0'>
      <Icon className='size-5 shrink-0 text-primary' />
      <div className='min-w-0'>
        <p className='text-xs text-muted-foreground'>{label}</p>
        <p className='mt-1 text-xl font-semibold tabular-nums'>{value}</p>
        <p className='truncate text-[11px] text-muted-foreground'>{hint}</p>
      </div>
    </div>
  );
}
function PageTitle() {
  return (
    <div className='flex items-center gap-2'>
      <Layers3 className='size-5 text-primary' />
      <h1 className='text-2xl font-semibold tracking-tight'>服务总览</h1>
    </div>
  );
}
