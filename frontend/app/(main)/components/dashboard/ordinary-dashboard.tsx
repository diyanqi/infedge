'use client';

import Link from 'next/link';
import { useQueries, useQuery, useQueryClient } from '@tanstack/react-query';
import type { LucideIcon } from 'lucide-react';
import {
  ArrowRight,
  CheckCircle2,
  CircleAlert,
  Cloud,
  FileText,
  Globe,
  Layers3,
  MapPin,
  Plus,
  RefreshCw,
  Rocket,
  ShieldCheck,
  Server,
  Settings2,
} from 'lucide-react';

import { EmptyStateWithBorder } from '@/components/layout/empty';
import { ErrorInline } from '@/components/layout/error';
import { LoadingStateWithBorder } from '@/components/layout/loading';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { useUser } from '@/contexts/user-context';
import { CustomService } from '@/lib/services/custom';
import type { ResourceRoute, ResourceZone } from '@/lib/services/custom';
import { PagesService } from '@/lib/services/openflare';

const resourceQueryKeys = {
  zones: ['custom', 'zones'],
  origins: ['custom', 'origins'],
  routes: ['custom', 'routes'],
  pages: ['openflare', 'pages', 'projects'],
  policies: ['custom', 'policies'],
} as const;

function getDomainCount(zones: ResourceZone[]) {
  return zones.reduce((count, zone) => count + (zone.domain_count ?? 0), 0);
}

function getStatus(route: ResourceRoute) {
  if (!route.enabled) return { label: '已停用', variant: 'outline' as const };
  if (!route.zone_domains.length) {
    return { label: '待配置域名', variant: 'outline' as const };
  }
  return { label: '运行中', variant: 'secondary' as const };
}

function Metric({
  icon: Icon,
  label,
  value,
  hint,
}: {
  icon: LucideIcon;
  label: string;
  value: number;
  hint: string;
}) {
  return (
    <div className='flex min-w-0 items-center gap-3 border-b border-dashed px-4 py-4 last:border-b-0 sm:border-b-0 sm:border-r sm:last:border-r-0'>
      <div className='flex size-9 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary'>
        <Icon className='size-4' />
      </div>
      <div className='min-w-0'>
        <p className='text-xs text-muted-foreground'>{label}</p>
        <p className='mt-1 text-xl font-semibold tabular-nums'>{value}</p>
        <p className='truncate text-[11px] text-muted-foreground'>{hint}</p>
      </div>
    </div>
  );
}

function QuickAction({
  href,
  icon: Icon,
  title,
  description,
}: {
  href: string;
  icon: LucideIcon;
  title: string;
  description: string;
}) {
  return (
    <Link
      href={href}
      className='group flex items-center gap-3 border-b border-dashed px-4 py-4 last:border-b-0 hover:bg-muted/40 sm:border-b-0 sm:border-r sm:last:border-r-0'
    >
      <div className='flex size-9 shrink-0 items-center justify-center rounded-md border bg-background text-primary'>
        <Icon className='size-4' />
      </div>
      <div className='min-w-0 flex-1'>
        <p className='text-sm font-medium'>{title}</p>
        <p className='mt-1 truncate text-xs text-muted-foreground'>
          {description}
        </p>
      </div>
      <ArrowRight className='size-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5' />
    </Link>
  );
}

function SetupStep({
  done,
  title,
  description,
}: {
  done: boolean;
  title: string;
  description: string;
}) {
  return (
    <div className='flex items-start gap-3'>
      {done ? (
        <CheckCircle2 className='mt-0.5 size-4 shrink-0 text-emerald-600' />
      ) : (
        <CircleAlert className='mt-0.5 size-4 shrink-0 text-amber-600' />
      )}
      <div className='min-w-0'>
        <p className='text-sm font-medium'>{title}</p>
        <p className='mt-1 text-xs leading-5 text-muted-foreground'>
          {description}
        </p>
      </div>
    </div>
  );
}

function SitesList({ routes }: { routes: ResourceRoute[] }) {
  if (!routes.length) {
    return (
      <EmptyStateWithBorder
        icon={Globe}
        title='还没有 CDN 站点'
        description='添加根域、源站和域名后，就可以创建你的第一个 CDN 站点。'
      />
    );
  }

  return (
    <div className='divide-y divide-dashed border-y'>
      {routes.slice(0, 6).map((route) => {
        const status = getStatus(route);
        const domains = route.zone_domains.map((domain) => domain.domain);
        return (
          <div
            key={route.id}
            className='flex flex-col gap-3 py-4 sm:flex-row sm:items-center sm:justify-between'
          >
            <div className='flex min-w-0 items-start gap-3'>
              <div className='mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-md bg-muted'>
                <Cloud className='size-4 text-primary' />
              </div>
              <div className='min-w-0'>
                <div className='flex flex-wrap items-center gap-2'>
                  <p className='truncate text-sm font-medium'>
                    {route.site_name || domains[0] || `站点 #${route.id}`}
                  </p>
                  <Badge variant={status.variant}>{status.label}</Badge>
                </div>
                <p className='mt-1 truncate text-xs text-muted-foreground'>
                  {domains.length ? domains.join('、') : '尚未绑定域名'}
                </p>
              </div>
            </div>
            <Button variant='outline' size='sm' asChild>
              <Link href='/resources'>
                <Settings2 data-icon='inline-start' />
                管理站点
              </Link>
            </Button>
          </div>
        );
      })}
    </div>
  );
}

export function OrdinaryDashboard() {
  const { user } = useUser();
  const queryClient = useQueryClient();
  const zonesQuery = useQuery({
    queryKey: resourceQueryKeys.zones,
    queryFn: () => CustomService.listZones(),
  });
  const originsQuery = useQuery({
    queryKey: resourceQueryKeys.origins,
    queryFn: () => CustomService.listOrigins(),
  });
  const routesQuery = useQuery({
    queryKey: resourceQueryKeys.routes,
    queryFn: () => CustomService.listRoutes(),
  });
  const pagesQuery = useQuery({
    queryKey: resourceQueryKeys.pages,
    queryFn: () => PagesService.listProjects(),
  });
  const policiesQuery = useQuery({
    queryKey: resourceQueryKeys.policies,
    queryFn: () => CustomService.listPolicies(),
  });
  const zoneDetailQueries = useQueries({
    queries: (zonesQuery.data ?? []).map((zone) => ({
      queryKey: ['custom', 'zone', zone.id],
      queryFn: () => CustomService.getZone(zone.id),
    })),
  });

  const isLoading =
    zonesQuery.isLoading ||
    originsQuery.isLoading ||
    routesQuery.isLoading ||
    pagesQuery.isLoading;
  const firstError = [zonesQuery, originsQuery, routesQuery, pagesQuery].find(
    (query) => query.isError,
  );
  const zones = zonesQuery.data ?? [];
  const routes = routesQuery.data ?? [];
  const domains = zoneDetailQueries.flatMap(
    (query) => query.data?.domains ?? [],
  );
  const verifiedDomains = domains.filter(
    (domain) => domain.verification_status === 'verified',
  ).length;
  const configuredRoutes = routes.filter(
    (route) => route.enabled && route.zone_domains.length > 0,
  ).length;
  const setupDone = [
    zones.length > 0,
    originsQuery.data?.length ? originsQuery.data.length > 0 : false,
    configuredRoutes > 0,
  ];
  const setupProgress = Math.round(
    (setupDone.filter(Boolean).length / setupDone.length) * 100,
  );
  const displayName =
    user?.nickname || user?.username || user?.email?.split('@')[0] || '用户';

  const refresh = () => {
    Object.values(resourceQueryKeys).forEach((queryKey) => {
      void queryClient.invalidateQueries({ queryKey });
    });
    void zonesQuery.refetch();
  };

  if (isLoading) {
    return (
      <div className='w-full px-1 py-6'>
        <LoadingStateWithBorder
          title='加载 CDN 控制台'
          description='正在读取站点和接入状态...'
        />
      </div>
    );
  }

  if (firstError) {
    return (
      <div className='w-full space-y-6 px-1 py-6'>
        <div className='flex items-center gap-2'>
          <Layers3 className='size-5 text-primary' />
          <h1 className='text-2xl font-semibold tracking-tight'>控制台</h1>
        </div>
        <ErrorInline
          message={
            firstError.error instanceof Error
              ? firstError.error.message
              : 'CDN 资源加载失败'
          }
          onRetry={refresh}
        />
      </div>
    );
  }

  return (
    <div className='w-full space-y-6 px-1 py-6'>
      <div className='flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between'>
        <div>
          <div className='flex items-center gap-2'>
            <Layers3 className='size-5 text-primary' />
            <h1 className='text-2xl font-semibold tracking-tight'>控制台</h1>
          </div>
          <p className='mt-2 text-sm text-muted-foreground'>
            你好，{displayName}。从这里管理你的 CDN 站点和静态网站。
          </p>
        </div>
        <div className='flex items-center gap-2'>
          <Button variant='outline' size='sm' onClick={refresh}>
            <RefreshCw data-icon='inline-start' />
            刷新数据
          </Button>
          <Button size='sm' asChild>
            <Link href='/resources'>
              <Plus data-icon='inline-start' />
              添加站点
            </Link>
          </Button>
        </div>
      </div>

      <section className='overflow-hidden rounded-lg border bg-card'>
        <div className='grid sm:grid-cols-2 lg:grid-cols-4'>
          <Metric
            icon={Globe}
            label='CDN 站点'
            value={configuredRoutes}
            hint={`${routes.length} 条配置记录`}
          />
          <Metric
            icon={MapPin}
            label='托管域名'
            value={getDomainCount(zones) || domains.length}
            hint={`${verifiedDomains} 个已完成验证`}
          />
          <Metric
            icon={Server}
            label='源站地址'
            value={originsQuery.data?.length ?? 0}
            hint='可被多个站点复用'
          />
          <Metric
            icon={FileText}
            label='Pages 项目'
            value={pagesQuery.data?.length ?? 0}
            hint='静态网站部署项目'
          />
        </div>
      </section>

      <section className='overflow-hidden rounded-lg border bg-card'>
        <div className='border-b px-4 py-4'>
          <div className='flex items-center gap-2'>
            <Rocket className='size-4 text-primary' />
            <h2 className='text-base font-semibold'>快速操作</h2>
          </div>
          <p className='mt-1 text-xs text-muted-foreground'>
            选择一项开始配置，无需理解复杂的边缘节点设置。
          </p>
        </div>
        <div className='grid sm:grid-cols-2 lg:grid-cols-4'>
          <QuickAction
            href='/resources'
            icon={Plus}
            title='添加 CDN 站点'
            description='绑定域名和源站，开启加速'
          />
          <QuickAction
            href='/pages'
            icon={FileText}
            title='部署静态网站'
            description='上传或同步 Pages 部署包'
          />
          <QuickAction
            href='/resources'
            icon={ShieldCheck}
            title='检查安全配置'
            description='查看 HTTPS、WAF 和域名验证'
          />
          <QuickAction
            href='/plans'
            icon={Layers3}
            title='查看套餐'
            description='了解资源额度和流量权益'
          />
        </div>
      </section>

      <div className='grid gap-6 lg:grid-cols-[minmax(0,1.45fr)_minmax(280px,0.75fr)]'>
        <section className='rounded-lg border bg-card p-4 sm:p-5'>
          <div className='flex items-start justify-between gap-3'>
            <div>
              <h2 className='text-base font-semibold'>我的 CDN 站点</h2>
              <p className='mt-1 text-xs text-muted-foreground'>
                最近配置的站点状态
              </p>
            </div>
            <Button variant='ghost' size='sm' asChild>
              <Link href='/resources'>
                查看全部 <ArrowRight data-icon='inline-end' />
              </Link>
            </Button>
          </div>
          <div className='mt-4'>
            <SitesList routes={routes} />
          </div>
        </section>

        <section className='rounded-lg border bg-card p-4 sm:p-5'>
          <div className='flex items-start justify-between gap-3'>
            <div>
              <h2 className='text-base font-semibold'>接入进度</h2>
              <p className='mt-1 text-xs text-muted-foreground'>
                完成基础配置即可开始加速
              </p>
            </div>
            <span className='text-sm font-semibold tabular-nums'>
              {setupProgress}%
            </span>
          </div>
          <div className='mt-4 h-1.5 overflow-hidden rounded-full bg-muted'>
            <div
              className='h-full rounded-full bg-primary transition-all'
              style={{ width: `${setupProgress}%` }}
            />
          </div>
          <div className='mt-5 space-y-5'>
            <SetupStep
              done={setupDone[0]}
              title='添加根域'
              description='先添加你拥有的根域，建立资源隔离边界。'
            />
            <SetupStep
              done={setupDone[1]}
              title='配置源站'
              description='填写网站服务器地址，CDN 才能回源获取内容。'
            />
            <SetupStep
              done={setupDone[2]}
              title='创建 CDN 站点'
              description='将域名绑定到源站并完成 DNS 接入。'
            />
          </div>
          {policiesQuery.data?.cname ? (
            <div className='mt-5 border-t pt-4 text-xs text-muted-foreground'>
              接入 CNAME：
              <span className='font-mono text-foreground'>
                {policiesQuery.data.cname}
              </span>
            </div>
          ) : null}
        </section>
      </div>
    </div>
  );
}
