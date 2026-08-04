'use client';

import { useMemo, useState } from 'react';
import {
  useMutation,
  useQueries,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import {
  Globe,
  Link2,
  MapPin,
  Plus,
  Rocket,
  Route,
  Trash2,
} from 'lucide-react';
import { toast } from 'sonner';
import { CustomService } from '@/lib/services/custom';
import type { ResourceDomain } from '@/lib/services/custom';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';

export default function ResourcesPage() {
  const qc = useQueryClient();
  const [zoneDomain, setZoneDomain] = useState('');
  const [origin, setOrigin] = useState({ name: '', address: '', remark: '' });
  const [route, setRoute] = useState({ site: '', domain: '', origin: '' });
  const zones = useQuery({
    queryKey: ['custom', 'zones'],
    queryFn: () => CustomService.listZones(),
  });
  const origins = useQuery({
    queryKey: ['custom', 'origins'],
    queryFn: () => CustomService.listOrigins(),
  });
  const routes = useQuery({
    queryKey: ['custom', 'routes'],
    queryFn: () => CustomService.listRoutes(),
  });
  const domainQueries = useQueries({
    queries: (zones.data ?? []).map((zone) => ({
      queryKey: ['custom', 'zone', zone.id],
      queryFn: () => CustomService.getZone(zone.id),
    })),
  });
  const domains = useMemo(
    () =>
      domainQueries.reduce<Record<number, ResourceDomain[]>>(
        (map, query, index) => {
          const zoneID = zones.data?.[index]?.id;
          if (zoneID) map[zoneID] = query.data?.domains ?? [];
          return map;
        },
        {},
      ),
    [domainQueries, zones.data],
  );
  const refresh = () => {
    void qc.invalidateQueries({ queryKey: ['custom', 'zones'] });
    void qc.invalidateQueries({ queryKey: ['custom', 'origins'] });
    void qc.invalidateQueries({ queryKey: ['custom', 'routes'] });
  };
  const createZone = useMutation({
    mutationFn: () => CustomService.createZone(zoneDomain),
    onSuccess: () => {
      setZoneDomain('');
      refresh();
      toast.success('Zone 已创建');
    },
    onError: (e) => toast.error(e instanceof Error ? e.message : '创建失败'),
  });
  const createDomain = useMutation({
    mutationFn: ({ zoneId, domain }: { zoneId: number; domain: string }) =>
      CustomService.createDomain(zoneId, domain),
    onSuccess: () => {
      setZoneDomain('');
      void qc.invalidateQueries({ queryKey: ['custom', 'zones'] });
      toast.success('域名已添加');
    },
    onError: (e) => toast.error(e instanceof Error ? e.message : '添加失败'),
  });
  const createOrigin = useMutation({
    mutationFn: () => CustomService.createOrigin(origin),
    onSuccess: () => {
      setOrigin({ name: '', address: '', remark: '' });
      refresh();
      toast.success('源站已创建');
    },
    onError: (e) => toast.error(e instanceof Error ? e.message : '创建失败'),
  });
  const createRoute = useMutation({
    mutationFn: () =>
      CustomService.createRoute({
        site_name: route.site,
        zone_domain_ids: [Number(route.domain)],
        origin_id: Number(route.origin),
        origin_scheme: 'http',
        origin_port: '',
        origin_uri: '/',
        origin_host: '',
        origin_address: '',
        origin_url: '',
        upstreams: [],
        enabled: true,
        enable_https: false,
        redirect_http: false,
        cache_enabled: false,
        cache_policy: '',
        cache_rules: [],
        custom_headers: [],
        basic_auth_enabled: false,
        upstream_type: 'direct',
      }),
    onSuccess: () => {
      setRoute({ site: '', domain: '', origin: '' });
      refresh();
      toast.success('CDN 规则已创建');
    },
    onError: (e) => toast.error(e instanceof Error ? e.message : '创建失败'),
  });
  const publish = useMutation({
    mutationFn: () => CustomService.publish(),
    onSuccess: (data) => toast.success('已发布版本 ' + data.version),
    onError: (e) => toast.error(e instanceof Error ? e.message : '发布失败'),
  });
  const remove = (kind: 'zone' | 'origin' | 'route', id: number) => {
    const task =
      kind === 'zone'
        ? CustomService.deleteZone(id)
        : kind === 'origin'
          ? CustomService.deleteOrigin(id)
          : CustomService.deleteRoute(id);
    void task
      .then(() => {
        refresh();
        toast.success('已删除');
      })
      .catch((e) => toast.error(e instanceof Error ? e.message : '删除失败'));
  };
  const allDomains = Object.values(domains).flat();

  return (
    <div className='w-full space-y-6 py-6 px-1'>
      <div className='flex flex-wrap items-center justify-between gap-3'>
        <div className='flex items-center gap-2'>
          <Globe className='size-5 text-primary' />
          <h1 className='text-2xl font-semibold tracking-tight'>我的资源</h1>
        </div>
        <Button onClick={() => publish.mutate()} disabled={publish.isPending}>
          <Rocket className='mr-2 size-4' />
          发布全站版本
        </Button>
      </div>
      <Tabs defaultValue='zones'>
        <TabsList>
          <TabsTrigger value='zones'>
            <Globe className='mr-2 size-4' />
            域名
          </TabsTrigger>
          <TabsTrigger value='origins'>
            <MapPin className='mr-2 size-4' />
            源站
          </TabsTrigger>
          <TabsTrigger value='routes'>
            <Route className='mr-2 size-4' />
            CDN 规则
          </TabsTrigger>
        </TabsList>
        <TabsContent value='zones' className='space-y-4'>
          <Card>
            <CardHeader>
              <CardTitle className='text-base'>添加 Zone</CardTitle>
            </CardHeader>
            <CardContent className='flex flex-col gap-3 sm:flex-row'>
              <div className='flex-1'>
                <Label htmlFor='zone-domain'>根域</Label>
                <Input
                  id='zone-domain'
                  value={zoneDomain}
                  onChange={(e) => setZoneDomain(e.target.value)}
                  placeholder='example.com'
                />
              </div>
              <Button
                className='self-end'
                onClick={() => createZone.mutate()}
                disabled={!zoneDomain || createZone.isPending}
              >
                <Plus className='mr-2 size-4' />
                创建 Zone
              </Button>
            </CardContent>
          </Card>
          <div className='grid gap-3 md:grid-cols-2'>
            {(zones.data ?? []).map((zone) => (
              <Card key={zone.id} className='border-dashed'>
                <CardHeader className='pb-3'>
                  <div className='flex items-center justify-between'>
                    <CardTitle className='text-base'>{zone.domain}</CardTitle>
                    <Button
                      variant='ghost'
                      size='icon'
                      onClick={() => remove('zone', zone.id)}
                      aria-label='删除 Zone'
                    >
                      <Trash2 className='size-4' />
                    </Button>
                  </div>
                </CardHeader>
                <CardContent className='space-y-3'>
                  {(domains[zone.id] ?? []).map((domain) => (
                    <div
                      key={domain.id}
                      className='flex items-center justify-between border-b border-dashed py-1 text-sm'
                    >
                      <span>{domain.domain}</span>
                      <Badge variant='outline'>
                        {domain.proxy_route_id ? '已绑定' : '未绑定'}
                      </Badge>
                    </div>
                  ))}
                  <div className='flex gap-2'>
                    <Input
                      placeholder='www.example.com'
                      value={zoneDomain}
                      onChange={(e) => setZoneDomain(e.target.value)}
                    />
                    <Button
                      size='icon'
                      onClick={() =>
                        createDomain.mutate({
                          zoneId: zone.id,
                          domain: zoneDomain,
                        })
                      }
                      disabled={!zoneDomain}
                      aria-label='添加域名'
                    >
                      <Link2 className='size-4' />
                    </Button>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>
        <TabsContent value='origins' className='space-y-4'>
          <Card>
            <CardHeader>
              <CardTitle className='text-base'>添加源站</CardTitle>
            </CardHeader>
            <CardContent className='grid gap-3 sm:grid-cols-4'>
              <Input
                placeholder='名称'
                value={origin.name}
                onChange={(e) => setOrigin({ ...origin, name: e.target.value })}
              />
              <Input
                placeholder='源站地址，如 10.0.0.8'
                value={origin.address}
                onChange={(e) =>
                  setOrigin({ ...origin, address: e.target.value })
                }
              />
              <Input
                placeholder='备注'
                value={origin.remark}
                onChange={(e) =>
                  setOrigin({ ...origin, remark: e.target.value })
                }
              />
              <Button
                onClick={() => createOrigin.mutate()}
                disabled={!origin.address}
              >
                <Plus className='mr-2 size-4' />
                创建源站
              </Button>
            </CardContent>
          </Card>
          <div className='grid gap-3 md:grid-cols-2'>
            {(origins.data ?? []).map((item) => (
              <Card key={item.id} className='border-dashed'>
                <CardContent className='flex items-center justify-between gap-3 pt-6'>
                  <div>
                    <p className='font-medium'>{item.name}</p>
                    <p className='text-sm text-muted-foreground'>
                      {item.address}
                    </p>
                  </div>
                  <Button
                    variant='ghost'
                    size='icon'
                    onClick={() => remove('origin', item.id)}
                    aria-label='删除源站'
                  >
                    <Trash2 className='size-4' />
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>
        <TabsContent value='routes' className='space-y-4'>
          <Card>
            <CardHeader>
              <CardTitle className='text-base'>添加 CDN 规则</CardTitle>
            </CardHeader>
            <CardContent className='grid gap-3 sm:grid-cols-4'>
              <Input
                placeholder='站点名称'
                value={route.site}
                onChange={(e) => setRoute({ ...route, site: e.target.value })}
              />
              <Select
                value={route.domain}
                onValueChange={(value) => setRoute({ ...route, domain: value })}
              >
                <SelectTrigger>
                  <SelectValue placeholder='选择域名' />
                </SelectTrigger>
                <SelectContent>
                  {allDomains.map((domain) => (
                    <SelectItem key={domain.id} value={String(domain.id)}>
                      {domain.domain}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Select
                value={route.origin}
                onValueChange={(value) => setRoute({ ...route, origin: value })}
              >
                <SelectTrigger>
                  <SelectValue placeholder='选择源站' />
                </SelectTrigger>
                <SelectContent>
                  {(origins.data ?? []).map((item) => (
                    <SelectItem key={item.id} value={String(item.id)}>
                      {item.name} · {item.address}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Button
                onClick={() => createRoute.mutate()}
                disabled={!route.site || !route.domain || !route.origin}
              >
                <Plus className='mr-2 size-4' />
                创建规则
              </Button>
            </CardContent>
          </Card>
          <div className='grid gap-3 md:grid-cols-2'>
            {(routes.data ?? []).map((item) => (
              <Card key={item.id} className='border-dashed'>
                <CardContent className='flex items-center justify-between gap-3 pt-6'>
                  <div>
                    <p className='font-medium'>{item.site_name}</p>
                    <p className='text-sm text-muted-foreground'>
                      {item.zone_domains?.map((d) => d.domain).join(', ') ||
                        item.origin_url}
                    </p>
                  </div>
                  <Button
                    variant='ghost'
                    size='icon'
                    onClick={() => remove('route', item.id)}
                    aria-label='删除规则'
                  >
                    <Trash2 className='size-4' />
                  </Button>
                </CardContent>
              </Card>
            ))}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
