// 普通用户站点控制台：Tab 化管理源站、HTTPS、缓存与安全配置。
'use client';

import Link from 'next/link';
import { useCallback, useMemo } from 'react';
import { usePathname, useRouter, useSearchParams } from 'next/navigation';
import { ArrowLeft, Globe, Rocket, Settings2 } from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { LoadingStateWithBorder } from '@/components/layout/loading';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

import { useSiteConsole } from './use-site-console';
import { CacheTab } from './tabs/cache-tab';
import { DomainsTab } from './tabs/domains-tab';
import { HttpsTab } from './tabs/https-tab';
import { OriginsTab } from './tabs/origins-tab';
import { OverviewTab } from './tabs/overview-tab';
import { SecurityTab } from './tabs/security-tab';

const siteTabs = [
  'overview',
  'domains',
  'origins',
  'https',
  'cache',
  'security',
] as const;
type SiteTab = (typeof siteTabs)[number];

function getSiteTab(value: string | null | undefined): SiteTab {
  return siteTabs.includes(value as SiteTab) ? (value as SiteTab) : 'overview';
}

export default function ResourceDetailPage() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const activeTab = useMemo(
    () => getSiteTab(searchParams.get('tab')),
    [searchParams],
  );
  const console = useSiteConsole();

  const setActiveTab = useCallback(
    (tab: string) => {
      const next = getSiteTab(tab);
      const params = new URLSearchParams(searchParams.toString());
      if (next === 'overview') {
        params.delete('tab');
      } else {
        params.set('tab', next);
      }
      const query = params.toString();
      router.replace(query ? `${pathname}?${query}` : pathname, {
        scroll: false,
      });
    },
    [pathname, router, searchParams],
  );

  const {
    route,
    selectedDomain,
    selectedZone,
    domainVerified,
    loading,
    siteName,
    originId,
    upstreamType,
    saveRoute,
    publish,
  } = console;

  if (loading) {
    return (
      <div className='w-full px-1 py-6'>
        <LoadingStateWithBorder
          title='加载站点控制台'
          description='正在读取站点配置...'
        />
      </div>
    );
  }

  const domain = selectedDomain?.domain ?? route?.site_name ?? '未命名站点';
  const running = Boolean(
    route?.enabled && route?.zone_domains?.length && domainVerified,
  );
  const status = running ? '运行中' : route ? '待配置' : '待验证';
  const canSave =
    Boolean(selectedDomain) &&
    Boolean(domainVerified) &&
    Boolean(siteName.trim()) &&
    !(upstreamType === 'direct' && !originId);

  return (
    <div className='w-full space-y-5 px-1 py-6'>
      <div className='flex flex-wrap items-center justify-between gap-3'>
        <div className='flex items-center gap-2'>
          <Button variant='ghost' size='icon' asChild>
            <Link href='/resources' aria-label='返回网站列表'>
              <ArrowLeft />
            </Link>
          </Button>
          <Globe className='size-5 text-primary' />
          <h1 className='text-2xl font-semibold tracking-tight'>站点控制台</h1>
        </div>
        <div className='flex items-center gap-2'>
          <Button
            onClick={() => saveRoute.mutate()}
            disabled={!canSave || saveRoute.isPending}
          >
            <Settings2 data-icon='inline-start' />
            {saveRoute.isPending
              ? '保存中…'
              : route
                ? '保存配置'
                : '创建并保存'}
          </Button>
          <Button onClick={() => publish.mutate()} disabled={publish.isPending}>
            <Rocket data-icon='inline-start' />
            {publish.isPending ? '部署中…' : '部署'}
          </Button>
        </div>
      </div>

      <div className='flex flex-wrap items-center gap-3 rounded-lg border bg-card p-4'>
        <div className='flex size-10 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary'>
          <Globe className='size-5' />
        </div>
        <div className='min-w-0 flex-1'>
          <div className='flex flex-wrap items-center gap-2'>
            <h2 className='truncate text-lg font-semibold'>{domain}</h2>
            <Badge variant={running ? 'secondary' : 'outline'}>{status}</Badge>
          </div>
          <p className='mt-1 text-xs text-muted-foreground'>
            {selectedZone ? `域名分组：${selectedZone.domain}` : ''}
            {route ? ` · 站点 ID ${route.id}` : ''}
            {!domainVerified && selectedDomain
              ? ' · 请先完成 TXT 所有权验证'
              : ''}
          </p>
        </div>
      </div>

      {!route && !selectedDomain ? (
        <Card>
          <CardContent className='py-12 text-center text-sm text-muted-foreground'>
            找不到这个站点，可能已被删除或不属于当前账号。
          </CardContent>
        </Card>
      ) : (
        <Tabs value={activeTab} onValueChange={setActiveTab} className='w-full'>
          <TabsList variant='line' className='mb-6 inline-flex w-fit gap-6'>
            <TabsTrigger value='overview'>概览</TabsTrigger>
            <TabsTrigger value='domains'>域名</TabsTrigger>
            <TabsTrigger value='origins'>源站</TabsTrigger>
            <TabsTrigger value='https'>HTTPS / 证书</TabsTrigger>
            <TabsTrigger value='cache'>缓存</TabsTrigger>
            <TabsTrigger value='security'>安全防护</TabsTrigger>
          </TabsList>
          <TabsContent value='overview'>
            <OverviewTab console={console} />
          </TabsContent>
          <TabsContent value='domains'>
            <DomainsTab console={console} />
          </TabsContent>
          <TabsContent value='origins'>
            <OriginsTab console={console} />
          </TabsContent>
          <TabsContent value='https'>
            <HttpsTab console={console} />
          </TabsContent>
          <TabsContent value='cache'>
            <CacheTab console={console} />
          </TabsContent>
          <TabsContent value='security'>
            <SecurityTab console={console} />
          </TabsContent>
        </Tabs>
      )}
    </div>
  );
}
