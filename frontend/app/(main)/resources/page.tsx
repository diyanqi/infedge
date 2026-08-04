'use client';

import { useEffect, useMemo, useState } from 'react';
import { useMutation, useQueries, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  CheckCircle2,
  Globe,
  Link2,
  LockKeyhole,
  MapPin,
  Plus,
  Rocket,
  Route,
  Settings2,
  ShieldAlert,
  ShieldCheck,
  Trash2,
} from 'lucide-react';
import { toast } from 'sonner';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { CustomService } from '@/lib/services/custom';
import type { ResourceDomain, ResourceRoute } from '@/lib/services/custom';

const emptyOrigin = { name: '', address: '', remark: '' };

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

function routePayload(
  route: ResourceRoute | undefined,
  input: {
    siteName: string;
    domainId: number;
    originId: string;
    enableHttps: boolean;
    redirectHttp: boolean;
    limitRate: string;
    limitReqPerIP: string;
  },
) {
  return {
    site_name: input.siteName,
    zone_domain_ids: [input.domainId],
    origin_id: input.originId ? Number(input.originId) : null,
    origin_url: route?.origin_url ?? '',
    origin_scheme: 'http',
    origin_address: '',
    origin_port: '',
    origin_uri: '/',
    origin_host: route?.origin_host ?? '',
    upstreams: route?.upstream_list ?? [],
    enabled: route?.enabled ?? true,
    enable_https: input.enableHttps,
    redirect_http: input.redirectHttp,
    limit_conn_per_server: route?.limit_conn_per_server ?? 0,
    limit_conn_per_ip: route?.limit_conn_per_ip ?? 0,
    limit_rate: input.limitRate,
    limit_req_per_ip: input.limitReqPerIP,
    cache_enabled: route?.cache_enabled ?? false,
    cache_policy: route?.cache_policy ?? '',
    cache_rules: [],
    custom_headers: [],
    basic_auth_enabled: route?.basic_auth_enabled ?? false,
    basic_auth_username: route?.basic_auth_username ?? '',
    basic_auth_password: '',
    upstream_type: route?.upstream_type ?? 'direct',
    pages_project_id: route?.pages_project_id ?? null,
  };
}

export default function ResourcesPage() {
  const queryClient = useQueryClient();
  const [selectedZoneId, setSelectedZoneId] = useState<number | null>(null);
  const [selectedDomainId, setSelectedDomainId] = useState<number | null>(null);
  const [rootDomain, setRootDomain] = useState('');
  const [claimsOwnership, setClaimsOwnership] = useState(false);
  const [childDomain, setChildDomain] = useState('');
  const [childCertId, setChildCertId] = useState('');
  const [origin, setOrigin] = useState(emptyOrigin);
  const [siteName, setSiteName] = useState('');
  const [originId, setOriginId] = useState('');
  const [enableHttps, setEnableHttps] = useState(false);
  const [redirectHttp, setRedirectHttp] = useState(false);
  const [limitRate, setLimitRate] = useState('');
  const [limitReqPerIP, setLimitReqPerIP] = useState('');
  const [wafName, setWafName] = useState('');
  const [wafDraft, setWafDraft] = useState<number[]>([]);

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
  const certificates = useQuery({
    queryKey: ['custom', 'certificates'],
    queryFn: () => CustomService.listCertificates(),
  });
  const policies = useQuery({
    queryKey: ['custom', 'policies'],
    queryFn: () => CustomService.listPolicies(),
  });
  const wafRules = useQuery({
    queryKey: ['custom', 'waf-rules'],
    queryFn: () => CustomService.listWafRules(),
  });
  const domainQueries = useQueries({
    queries: (zones.data ?? []).map((zone) => ({
      queryKey: ['custom', 'zone', zone.id],
      queryFn: () => CustomService.getZone(zone.id),
    })),
  });
  const domains = useMemo(
    () =>
      domainQueries.reduce<Record<number, ResourceDomain[]>>((result, query, index) => {
        const zoneId = zones.data?.[index]?.id;
        if (zoneId) result[zoneId] = query.data?.domains ?? [];
        return result;
      }, {}),
    [domainQueries, zones.data],
  );
  const selectedZone = zones.data?.find((zone) => zone.id === selectedZoneId);
  const selectedDomains = selectedZone ? domains[selectedZone.id] ?? [] : [];
  const selectedDomain = selectedDomains.find((domain) => domain.id === selectedDomainId);
  const selectedRoute = routes.data?.find((route) =>
    route.zone_domain_ids.includes(selectedDomainId ?? -1),
  );
  const routeWaf = useQuery({
    queryKey: ['custom', 'route-waf', selectedRoute?.id],
    queryFn: () => CustomService.getRouteWaf(selectedRoute!.id),
    enabled: Boolean(selectedRoute),
  });

  useEffect(() => {
    if (!zones.data?.length) {
      setSelectedZoneId(null);
      setSelectedDomainId(null);
      return;
    }
    const zone = zones.data.find((item) => item.id === selectedZoneId) ?? zones.data[0];
    if (zone.id !== selectedZoneId) setSelectedZoneId(zone.id);
    const zoneDomains = domains[zone.id] ?? [];
    if (!zoneDomains.some((item) => item.id === selectedDomainId)) {
      setSelectedDomainId(zoneDomains[0]?.id ?? null);
    }
  }, [domains, selectedDomainId, selectedZoneId, zones.data]);

  useEffect(() => {
    setSiteName(selectedRoute?.site_name ?? '');
    setOriginId(selectedRoute?.origin_id ? String(selectedRoute.origin_id) : '');
    setEnableHttps(selectedRoute?.enable_https ?? false);
    setRedirectHttp(selectedRoute?.redirect_http ?? false);
    setLimitRate(selectedRoute?.limit_rate ?? '');
    setLimitReqPerIP(selectedRoute?.limit_req_per_ip ?? '');
  }, [selectedRoute]);

  useEffect(() => {
    setChildCertId(selectedDomain?.cert_id ? String(selectedDomain.cert_id) : '');
  }, [selectedDomain]);

  useEffect(() => {
    setWafDraft(routeWaf.data?.applied_ids ?? []);
  }, [routeWaf.data]);

  const refresh = () => {
    void queryClient.invalidateQueries({ queryKey: ['custom', 'zones'] });
    void queryClient.invalidateQueries({ queryKey: ['custom', 'origins'] });
    void queryClient.invalidateQueries({ queryKey: ['custom', 'routes'] });
    void queryClient.invalidateQueries({ queryKey: ['custom', 'certificates'] });
    void queryClient.invalidateQueries({ queryKey: ['custom', 'waf-rules'] });
  };

  const createZone = useMutation({
    mutationFn: () => CustomService.createZone(rootDomain, claimsOwnership),
    onSuccess: () => {
      setRootDomain('');
      setClaimsOwnership(false);
      refresh();
      toast.success('根域已创建，请按提示完成 DNS TXT 验证');
    },
    onError: (error) => toast.error(errorMessage(error, '创建根域失败')),
  });
  const createDomain = useMutation({
    mutationFn: () =>
      CustomService.createDomain(
        selectedZone!.id,
        childDomain,
        childCertId ? Number(childCertId) : null,
      ),
    onSuccess: () => {
      setChildDomain('');
      setChildCertId('');
      refresh();
      toast.success('子域已添加');
    },
    onError: (error) => toast.error(errorMessage(error, '添加子域失败')),
  });
  const updateDomain = useMutation({
    mutationFn: (domain: ResourceDomain) =>
      CustomService.updateDomain(selectedZone!.id, domain.id, {
        domain: domain.domain,
        cert_id: childCertId ? Number(childCertId) : null,
      }),
    onSuccess: () => {
      refresh();
      toast.success('TLS 证书配置已保存');
    },
    onError: (error) => toast.error(errorMessage(error, '保存证书配置失败')),
  });
  const createOrigin = useMutation({
    mutationFn: () => CustomService.createOrigin(origin),
    onSuccess: () => {
      setOrigin(emptyOrigin);
      refresh();
      toast.success('源站已创建');
    },
    onError: (error) => toast.error(errorMessage(error, '创建源站失败')),
  });
  const saveRoute = useMutation({
    mutationFn: () => {
      if (!selectedDomain) throw new Error('请先选择子域');
      return selectedRoute
        ? CustomService.updateRoute(
            selectedRoute.id,
            routePayload(selectedRoute, {
              siteName,
              domainId: selectedDomain.id,
              originId,
              enableHttps,
              redirectHttp,
              limitRate,
              limitReqPerIP,
            }),
          )
        : CustomService.createRoute(
            routePayload(undefined, {
              siteName,
              domainId: selectedDomain.id,
              originId,
              enableHttps,
              redirectHttp,
              limitRate,
              limitReqPerIP,
            }),
          );
    },
    onSuccess: () => {
      refresh();
      toast.success('子域配置已保存');
    },
    onError: (error) => toast.error(errorMessage(error, '保存子域配置失败')),
  });
  const publish = useMutation({
    mutationFn: () => CustomService.publish(),
    onSuccess: (data) => toast.success(`已部署版本 ${data.version}`),
    onError: (error) => toast.error(errorMessage(error, '部署失败')),
  });
  const verifyZone = useMutation({
    mutationFn: (id: number) => CustomService.verifyZone(id),
    onSuccess: () => {
      refresh();
      toast.success('根域验证成功，子域可继承验证状态');
    },
    onError: (error) => toast.error(errorMessage(error, '根域验证失败')),
  });
  const verifyDomain = useMutation({
    mutationFn: ({ zoneId, domainId }: { zoneId: number; domainId: number }) =>
      CustomService.verifyDomain(zoneId, domainId),
    onSuccess: () => {
      refresh();
      toast.success('子域验证成功');
    },
    onError: (error) => toast.error(errorMessage(error, '子域验证失败')),
  });
  const createWaf = useMutation({
    mutationFn: () => {
      if (!selectedDomain) throw new Error('请先选择子域');
      return CustomService.createWafRule({ name: wafName, host: selectedDomain.domain });
    },
    onSuccess: () => {
      setWafName('');
      refresh();
      toast.success('防火墙规则已创建');
    },
    onError: (error) => toast.error(errorMessage(error, '创建防火墙规则失败')),
  });
  const saveWaf = useMutation({
    mutationFn: () => {
      if (!selectedRoute) throw new Error('请先保存子域配置');
      return CustomService.updateRouteWaf(selectedRoute.id, wafDraft);
    },
    onSuccess: () => {
      void routeWaf.refetch();
      toast.success('防火墙绑定已保存');
    },
    onError: (error) => toast.error(errorMessage(error, '保存防火墙绑定失败')),
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
      .catch((error) => toast.error(errorMessage(error, '删除失败')));
  };

  return (
    <div className='w-full flex flex-col gap-6 py-6 px-1'>
      <div className='flex flex-wrap items-center justify-between gap-3'>
        <div className='flex items-center gap-2'>
          <Globe className='size-5 text-primary' />
          <h1 className='text-2xl font-semibold tracking-tight'>我的资源</h1>
        </div>
        <Button onClick={() => publish.mutate()} disabled={publish.isPending}>
          <Rocket data-icon='inline-start' />
          部署
        </Button>
      </div>

      <Alert variant='destructive'>
        <ShieldAlert />
        <AlertTitle>域名接入要求</AlertTitle>
        <AlertDescription>
          所有子域必须将 CNAME 指向 <strong>cname.edge.infvar.com</strong>。禁止自行优选 IP，发现后将封号且不可申诉；该 CNAME 已每小时进行全国拨测优选。
        </AlertDescription>
      </Alert>

      <Card>
        <CardHeader>
          <CardTitle className='text-base'>添加根域</CardTitle>
          <CardDescription>根域是资源隔离和所有权验证的第一级。</CardDescription>
        </CardHeader>
        <CardContent className='flex flex-col gap-3 lg:flex-row lg:items-end'>
          <div className='flex-1 flex flex-col gap-2'>
            <Label htmlFor='root-domain'>根域</Label>
            <Input
              id='root-domain'
              value={rootDomain}
              onChange={(event) => setRootDomain(event.target.value)}
              placeholder='example.com'
            />
          </div>
          <label className='flex items-center gap-2 pb-2 text-sm text-muted-foreground'>
            <Checkbox checked={claimsOwnership} onCheckedChange={(value) => setClaimsOwnership(value === true)} />
            我拥有根域全部所有权，子域可继承验证
          </label>
          <Button onClick={() => createZone.mutate()} disabled={!rootDomain || createZone.isPending}>
            <Plus data-icon='inline-start' />
            创建根域
          </Button>
        </CardContent>
      </Card>

      <div className='grid gap-6 lg:grid-cols-[220px_minmax(0,1fr)]'>
        <Card className='h-fit'>
          <CardHeader className='pb-3'>
            <CardTitle className='text-base'>根域</CardTitle>
          </CardHeader>
          <CardContent className='flex flex-col gap-1'>
            {(zones.data ?? []).map((zone) => (
              <Button
                key={zone.id}
                variant={zone.id === selectedZoneId ? 'secondary' : 'ghost'}
                className='justify-start'
                onClick={() => {
                  setSelectedZoneId(zone.id);
                  setSelectedDomainId((domains[zone.id] ?? [])[0]?.id ?? null);
                }}
              >
                <Globe data-icon='inline-start' />
                <span className='truncate'>{zone.domain}</span>
              </Button>
            ))}
            {!zones.data?.length && (
              <p className='px-2 py-4 text-sm text-muted-foreground'>暂无根域</p>
            )}
          </CardContent>
        </Card>

        <div className='flex min-w-0 flex-col gap-6'>
          {selectedZone ? (
            <Card>
              <CardHeader className='flex flex-row items-start justify-between gap-3'>
                <div className='flex flex-col gap-1'>
                  <CardTitle className='text-base'>{selectedZone.domain}</CardTitle>
                  <CardDescription>先选择子域，再进入该子域的配置。</CardDescription>
                </div>
                <div className='flex items-center gap-2'>
                  {selectedZone.verification_status === 'verified' ? (
                    <Badge variant='secondary'><CheckCircle2 data-icon='inline-start' />已验证</Badge>
                  ) : (
                    <Button variant='outline' size='sm' onClick={() => verifyZone.mutate(selectedZone.id)}>
                      验证根域
                    </Button>
                  )}
                  <Button variant='ghost' size='icon' onClick={() => remove('zone', selectedZone.id)} aria-label='删除根域'>
                    <Trash2 />
                  </Button>
                </div>
              </CardHeader>
              <CardContent className='flex flex-col gap-4'>
                <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-3'>
                  {selectedDomains.map((domain) => (
                    <Button
                      key={domain.id}
                      variant={domain.id === selectedDomainId ? 'secondary' : 'outline'}
                      className='h-auto min-h-16 justify-between px-4 py-3 text-left'
                      onClick={() => setSelectedDomainId(domain.id)}
                    >
                      <span className='flex min-w-0 flex-col gap-1'>
                        <span className='truncate'>{domain.domain}</span>
                        <span className='text-xs text-muted-foreground'>
                          {domain.proxy_route_id ? '已配置站点' : '尚未配置站点'}
                        </span>
                      </span>
                      {domain.verification_status === 'verified' ? <CheckCircle2 /> : <ShieldAlert />}
                    </Button>
                  ))}
                </div>
                <div className='flex flex-col gap-3 border-t pt-4'>
                  <div className='flex items-center gap-2 text-sm font-medium'>
                    <Link2 /> 添加子域
                  </div>
                  <div className='grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,220px)_auto]'>
                    <Input value={childDomain} onChange={(event) => setChildDomain(event.target.value)} placeholder='www.example.com' />
                    <Select value={childCertId} onValueChange={setChildCertId}>
                      <SelectTrigger><SelectValue placeholder='选择 HTTPS 证书（可选）' /></SelectTrigger>
                      <SelectContent>
                        {(certificates.data ?? []).map((certificate) => (
                          <SelectItem key={certificate.id} value={String(certificate.id)}>{certificate.name}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <Button onClick={() => createDomain.mutate()} disabled={!childDomain || createDomain.isPending}>
                      <Plus data-icon='inline-start' />添加
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ) : (
            <Card><CardContent className='py-12 text-center text-sm text-muted-foreground'>创建根域后，从左侧选择一个根域。</CardContent></Card>
          )}

          {selectedDomain && (
            <div className='flex flex-col gap-6'>
              <Card>
                <CardHeader>
                  <div className='flex flex-wrap items-start justify-between gap-3'>
                    <div className='flex flex-col gap-1'>
                      <CardTitle className='text-base'>{selectedDomain.domain}</CardTitle>
                      <CardDescription>子域配置只作用于当前 Host，不会影响其他子域。</CardDescription>
                    </div>
                    <div className='flex items-center gap-2'>
                      <Badge variant='outline'>{selectedDomain.verification_status === 'verified' ? 'TXT 已验证' : '待 TXT 验证'}</Badge>
                      {selectedDomain.verification_status !== 'verified' && !selectedZone?.claims_ownership && (
                        <Button size='sm' variant='outline' onClick={() => verifyDomain.mutate({ zoneId: selectedZone!.id, domainId: selectedDomain.id })}>验证子域</Button>
                      )}
                    </div>
                  </div>
                </CardHeader>
                <CardContent className='flex flex-col gap-5'>
                  <Alert>
                    <Link2 />
                    <AlertTitle>DNS 配置</AlertTitle>
                    <AlertDescription>请将 {selectedDomain.domain} 的 CNAME 指向 <strong>cname.edge.infvar.com</strong> 后再部署。</AlertDescription>
                  </Alert>
                  <div className='grid gap-4 md:grid-cols-2'>
                    <div className='flex flex-col gap-2'>
                      <Label>源站</Label>
                      <Select value={originId} onValueChange={setOriginId}>
                        <SelectTrigger><SelectValue placeholder='选择源站' /></SelectTrigger>
                        <SelectContent>
                          {(origins.data ?? []).map((item) => <SelectItem key={item.id} value={String(item.id)}>{item.name} · {item.address}</SelectItem>)}
                        </SelectContent>
                      </Select>
                    </div>
                    <div className='flex flex-col gap-2'>
                      <Label>HTTPS 与 TLS 证书</Label>
                      <div className='flex items-center gap-3'>
                        <Switch checked={enableHttps} onCheckedChange={setEnableHttps} aria-label='启用 HTTPS' />
                        <span className='text-sm'>{enableHttps ? 'HTTPS 已启用' : 'HTTPS 未启用'}</span>
                        <Badge variant='outline'>{selectedDomain.cert_id ? '已绑定证书' : '未绑定证书'}</Badge>
                      </div>
                      <Select value={childCertId || (selectedDomain.cert_id ? String(selectedDomain.cert_id) : '')} onValueChange={setChildCertId}>
                        <SelectTrigger><SelectValue placeholder='选择证书' /></SelectTrigger>
                        <SelectContent>
                          {(certificates.data ?? []).map((certificate) => <SelectItem key={certificate.id} value={String(certificate.id)}>{certificate.name}</SelectItem>)}
                        </SelectContent>
                      </Select>
                      <Button variant='outline' size='sm' onClick={() => updateDomain.mutate(selectedDomain)} disabled={updateDomain.isPending || !childCertId}>
                        <LockKeyhole data-icon='inline-start' />保存证书
                      </Button>
                    </div>
                  </div>
                  <div className='grid gap-3 md:grid-cols-2'>
                    <div className='flex flex-col gap-2'><Label htmlFor='site-name'>站点名称</Label><Input id='site-name' value={siteName} onChange={(event) => setSiteName(event.target.value)} placeholder='例如：官网' /></div>
                    <div className='flex items-center gap-3 pt-7'><Switch checked={redirectHttp} onCheckedChange={setRedirectHttp} disabled={!enableHttps} aria-label='HTTP 跳转 HTTPS' /><span className='text-sm'>HTTP 自动跳转 HTTPS</span></div>
                  </div>
                  <div className='flex flex-col gap-3 border-t pt-4'>
                    <div className='flex items-center gap-2 text-sm font-medium'><Settings2 />限流策略</div>
                    <div className='grid gap-3 md:grid-cols-2'><div className='flex flex-col gap-2'><Label htmlFor='limit-rate'>单连接限速</Label><Input id='limit-rate' value={limitRate} onChange={(event) => setLimitRate(event.target.value)} placeholder='留空使用全局默认' /></div><div className='flex flex-col gap-2'><Label htmlFor='limit-req'>单 IP 请求限流</Label><Input id='limit-req' value={limitReqPerIP} onChange={(event) => setLimitReqPerIP(event.target.value)} placeholder='留空使用全局默认' /></div></div>
                  </div>
                  <Button onClick={() => saveRoute.mutate()} disabled={!siteName || !originId || saveRoute.isPending}>
                    <Settings2 data-icon='inline-start' />{selectedRoute ? '保存子域配置' : '创建并保存子域配置'}
                  </Button>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <div className='flex flex-wrap items-start justify-between gap-3'>
                    <div className='flex flex-col gap-1'><CardTitle className='text-base'>防火墙</CardTitle><CardDescription>规则仅绑定当前子域；全局规则始终优先且只读。</CardDescription></div>
                    <ShieldCheck className='size-5 text-primary' />
                  </div>
                </CardHeader>
                <CardContent className='flex flex-col gap-5'>
                  <div className='flex flex-col gap-2'>
                    <Label htmlFor='waf-name'>创建当前子域规则</Label>
                    <div className='flex gap-2'><Input id='waf-name' value={wafName} onChange={(event) => setWafName(event.target.value)} placeholder={`例如：保护 ${selectedDomain.domain}`} /><Button onClick={() => createWaf.mutate()} disabled={!wafName || createWaf.isPending}><Plus data-icon='inline-start' />创建规则</Button></div>
                  </div>
                  {selectedRoute ? (
                    <div className='flex flex-col gap-3'>
                      <Label>绑定自定义规则</Label>
                      {(wafRules.data ?? []).filter((rule) => !rule.is_global).map((rule) => (
                        <label key={rule.id} className='flex items-center gap-3 rounded-md border px-3 py-2 text-sm'>
                          <Checkbox disabled={Boolean(rule.host && rule.host !== selectedDomain.domain)} checked={wafDraft.includes(rule.id)} onCheckedChange={(checked) => setWafDraft((current) => checked === true ? [...new Set([...current, rule.id])] : current.filter((id) => id !== rule.id))} />
                          <span className='flex min-w-0 flex-1 flex-col'><span>{rule.name}</span><span className='text-xs text-muted-foreground'>Host: {rule.host || '未指定'}</span></span>
                          <Badge variant={rule.host === selectedDomain.domain ? 'secondary' : 'outline'}>{rule.host === selectedDomain.domain ? '匹配当前子域' : '其他 Host'}</Badge>
                        </label>
                      ))}
                      <Button variant='outline' onClick={() => saveWaf.mutate()} disabled={saveWaf.isPending}>保存防火墙绑定</Button>
                    </div>
                  ) : <p className='text-sm text-muted-foreground'>请先保存源站和站点配置，再绑定防火墙规则。</p>}
                  <div className='flex flex-col gap-2 border-t pt-4'><Label>全局规则（只读）</Label>{(policies.data?.global_rules ?? []).map((rule) => <div key={rule.id} className='flex items-center gap-2 text-sm'><ShieldCheck className='size-4 text-primary' /><span>{rule.name}</span><Badge variant='secondary'>全局优先</Badge></div>)}{!policies.data?.global_rules?.length && <span className='text-sm text-muted-foreground'>暂无全局规则</span>}</div>
                </CardContent>
              </Card>
            </div>
          )}
        </div>
      </div>

      <Card>
        <CardHeader><CardTitle className='text-base'>源站资源</CardTitle><CardDescription>源站可被自己的多个子域配置复用。</CardDescription></CardHeader>
        <CardContent className='flex flex-col gap-4'>
          <div className='grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)_auto]'>
            <Input placeholder='名称' value={origin.name} onChange={(event) => setOrigin({ ...origin, name: event.target.value })} />
            <Input placeholder='源站地址，如 10.0.0.8' value={origin.address} onChange={(event) => setOrigin({ ...origin, address: event.target.value })} />
            <Input placeholder='备注' value={origin.remark} onChange={(event) => setOrigin({ ...origin, remark: event.target.value })} />
            <Button onClick={() => createOrigin.mutate()} disabled={!origin.address || createOrigin.isPending}><Plus data-icon='inline-start' />创建</Button>
          </div>
          <div className='grid gap-3 md:grid-cols-2'>{(origins.data ?? []).map((item) => <div key={item.id} className='flex items-center justify-between gap-3 rounded-md border px-3 py-3'><div><p className='font-medium'>{item.name}</p><p className='text-sm text-muted-foreground'>{item.address}</p></div><Button variant='ghost' size='icon' onClick={() => remove('origin', item.id)} aria-label='删除源站'><Trash2 /></Button></div>)}</div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader><CardTitle className='text-base'>全局策略（只读）</CardTitle><CardDescription>管理员配置的默认限流对所有用户生效，普通用户不能修改。</CardDescription></CardHeader>
        <CardContent className='grid gap-3 md:grid-cols-3'><div className='flex items-center gap-2 text-sm'><Route className='size-4 text-primary' />单连接默认限速：<strong>{policies.data?.default_limit_rate || '未设置'}</strong></div><div className='flex items-center gap-2 text-sm'><ShieldCheck className='size-4 text-primary' />单 IP 默认请求限流：<strong>{policies.data?.default_limit_req_per_ip || '未设置'}</strong></div><div className='flex items-center gap-2 text-sm'><MapPin className='size-4 text-primary' />固定 CNAME：<strong>{policies.data?.cname || 'cname.edge.infvar.com'}</strong></div></CardContent>
      </Card>
    </div>
  );
}
