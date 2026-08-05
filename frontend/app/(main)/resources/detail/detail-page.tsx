'use client';

import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import {
  ArrowLeft,
  ChevronRight,
  Globe,
  Route,
  Server,
  Settings2,
  ShieldCheck,
} from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { LoadingStateWithBorder } from '@/components/layout/loading';
import { ErrorInline } from '@/components/layout/error';
import { CustomService } from '@/lib/services/custom';

const sections = [
  { label: '站点概览', icon: Globe, anchor: 'overview' },
  { label: '域名服务', icon: Globe, anchor: 'domains' },
  { label: '源站配置', icon: Server, anchor: 'origins' },
  { label: '网站安全加速', icon: Route, anchor: 'acceleration' },
  { label: '安全防护', icon: ShieldCheck, anchor: 'security' },
] as const;

export default function ResourceDetailPage() {
  const params = useSearchParams();
  const id = Number(params.get('id'));
  const routes = useQuery({
    queryKey: ['custom', 'routes'],
    queryFn: () => CustomService.listRoutes(),
    enabled: Number.isFinite(id) && id > 0,
  });
  const route = routes.data?.find((item) => item.id === id);

  if (routes.isLoading)
    return (
      <div className='w-full px-1 py-6'>
        <LoadingStateWithBorder
          title='加载站点'
          description='正在读取站点配置...'
        />
      </div>
    );
  if (routes.isError)
    return (
      <div className='w-full space-y-6 px-1 py-6'>
        <Header />
        <ErrorInline
          message='站点加载失败'
          onRetry={() => void routes.refetch()}
        />
      </div>
    );
  if (!route)
    return (
      <div className='w-full space-y-6 px-1 py-6'>
        <Header />
        <Card>
          <CardContent className='py-12 text-center text-sm text-muted-foreground'>
            找不到这个站点，可能已被删除或不属于当前账号。
          </CardContent>
        </Card>
      </div>
    );

  const domain =
    route.zone_domains[0]?.domain || route.site_name || `站点 #${route.id}`;
  const status =
    route.enabled && route.zone_domains.length > 0
      ? '运行中'
      : route.enabled
        ? '待配置域名'
        : '已停用';

  return (
    <div className='w-full space-y-5 px-1 py-6'>
      <Header />
      <div className='flex flex-col gap-4 rounded-lg border bg-card p-4 sm:flex-row sm:items-center sm:justify-between'>
        <div className='flex min-w-0 items-center gap-3'>
          <div className='flex size-10 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary'>
            <Globe className='size-5' />
          </div>
          <div className='min-w-0'>
            <div className='flex flex-wrap items-center gap-2'>
              <h1 className='truncate text-lg font-semibold'>{domain}</h1>
              <Badge variant={status === '运行中' ? 'secondary' : 'outline'}>
                {status}
              </Badge>
            </div>
            <p className='mt-1 text-xs text-muted-foreground'>
              站点 ID {route.id} · {route.site_name || '未命名站点'}
            </p>
          </div>
        </div>
        <Button asChild>
          <Link href={`/resources/configure?route=${route.id}`}>
            <Settings2 data-icon='inline-start' />
            进入配置
          </Link>
        </Button>
      </div>
      <div className='grid gap-5 lg:grid-cols-[220px_minmax(0,1fr)]'>
        <aside className='h-fit rounded-lg border bg-card p-2'>
          <p className='px-3 py-2 text-xs font-medium text-muted-foreground'>
            站点设置
          </p>
          <nav className='space-y-1'>
            {sections.map(({ label, icon: Icon, anchor }, index) => (
              <a
                key={anchor}
                href={`#${anchor}`}
                className={`flex items-center gap-2 rounded-md px-3 py-2 text-sm ${index === 0 ? 'bg-muted font-medium text-foreground' : 'text-muted-foreground hover:bg-muted/60 hover:text-foreground'}`}
              >
                <Icon className='size-4' />
                <span>{label}</span>
                {index > 0 && <ChevronRight className='ml-auto size-3.5' />}
              </a>
            ))}
          </nav>
        </aside>
        <div className='min-w-0 space-y-5'>
          <section id='overview' className='scroll-mt-6'>
            <Card>
              <CardHeader>
                <CardTitle className='text-base'>站点概览</CardTitle>
              </CardHeader>
              <CardContent className='grid gap-4 sm:grid-cols-3'>
                <Info
                  label='访问域名'
                  value={
                    route.zone_domains.length
                      ? String(route.zone_domains.length)
                      : '未配置'
                  }
                />
                <Info
                  label='源站'
                  value={
                    route.origin_url ||
                    (route.origin_id ? `源站 #${route.origin_id}` : '未配置')
                  }
                />
                <Info
                  label='HTTPS'
                  value={route.enable_https ? '已启用' : '未启用'}
                />
              </CardContent>
            </Card>
          </section>
          <section id='domains' className='scroll-mt-6'>
            <SectionCard
              icon={Globe}
              title='域名服务'
              description={
                route.zone_domains.length
                  ? route.zone_domains.map((item) => item.domain).join('、')
                  : '还没有绑定域名'
              }
              href={`/resources/configure?route=${route.id}`}
            />
          </section>
          <section id='origins' className='scroll-mt-6'>
            <SectionCard
              icon={Server}
              title='源站配置'
              description={
                route.origin_url ||
                (route.origin_id
                  ? `已关联源站 #${route.origin_id}`
                  : '还没有配置源站')
              }
              href={`/resources/configure?route=${route.id}`}
            />
          </section>
          <section id='acceleration' className='scroll-mt-6'>
            <SectionCard
              icon={Route}
              title='网站安全加速'
              description={
                route.cache_enabled ? '边缘缓存已开启' : '边缘缓存使用默认策略'
              }
              href={`/resources/configure?route=${route.id}`}
            />
          </section>
          <section id='security' className='scroll-mt-6'>
            <SectionCard
              icon={ShieldCheck}
              title='安全防护'
              description='在配置页中管理 WAF 规则和站点级限流。'
              href={`/resources/configure?route=${route.id}`}
            />
          </section>
        </div>
      </div>
    </div>
  );
}

function Header() {
  return (
    <div className='flex items-center gap-2'>
      <Button variant='ghost' size='icon' asChild>
        <Link href='/resources' aria-label='返回网站列表'>
          <ArrowLeft />
        </Link>
      </Button>
      <Globe className='size-5 text-primary' />
      <h1 className='text-2xl font-semibold tracking-tight'>站点控制台</h1>
    </div>
  );
}
function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-md border bg-muted/20 p-3'>
      <p className='text-xs text-muted-foreground'>{label}</p>
      <p className='mt-2 truncate text-sm font-medium' title={value}>
        {value}
      </p>
    </div>
  );
}
function SectionCard({
  icon: Icon,
  title,
  description,
  href,
}: {
  icon: typeof Globe;
  title: string;
  description: string;
  href: string;
}) {
  return (
    <Card>
      <CardContent className='flex flex-col gap-3 p-4 sm:flex-row sm:items-center'>
        <Icon className='size-5 shrink-0 text-primary' />
        <div className='min-w-0 flex-1'>
          <p className='text-sm font-medium'>{title}</p>
          <p className='mt-1 truncate text-xs text-muted-foreground'>
            {description}
          </p>
        </div>
        <Button variant='outline' size='sm' asChild>
          <Link href={href}>
            配置
            <ChevronRight data-icon='inline-end' />
          </Link>
        </Button>
      </CardContent>
    </Card>
  );
}
