// Shared state, queries and mutations for the ordinary-user site console.
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import {
  useMutation,
  useQueries,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import { toast } from 'sonner';

import { CustomService } from '@/lib/services/custom';
import type {
  ResourceDomain,
  ResourceRoute,
  ResourceZone,
} from '@/lib/services/custom';

const emptyOrigin = { name: '', address: '', remark: '' };

export function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

export function certificateCoversDomain(
  certificate: { primary_domain?: string; other_domains?: string },
  domain: string,
) {
  const names =
    `${certificate.primary_domain ?? ''} ${certificate.other_domains ?? ''}`
      .split(/[\s,;]+/)
      .map((item) => item.trim().toLowerCase())
      .filter(Boolean);
  const normalized = domain.toLowerCase();
  return names.some(
    (name) =>
      name === normalized ||
      (name.startsWith('*.') && normalized.endsWith(name.slice(1))),
  );
}

function routePayload(
  route: ResourceRoute | undefined,
  input: {
    siteName: string;
    domainId: number;
    originId: string;
    originScheme: string;
    originPort: string;
    enableHttps: boolean;
    redirectHttp: boolean;
    limitRate: string;
    limitReqPerIP: string;
    upstreamsText: string;
    upstreamWeightsText: string;
    upstreamType: string;
    pagesProjectId: string;
    cacheEnabled: boolean;
    cachePolicy: string;
    cacheRulesText: string;
  },
) {
  return {
    site_name: input.siteName,
    zone_domain_ids: [input.domainId],
    origin_id: input.originId ? Number(input.originId) : null,
    origin_url: route?.origin_url ?? '',
    origin_scheme: input.originScheme,
    origin_address: '',
    origin_port: input.originPort,
    origin_uri: '/',
    origin_host: route?.origin_host ?? '',
    upstreams: input.upstreamsText
      .split(/\r?\n/)
      .map((item) => item.trim())
      .filter(Boolean),
    upstream_weights: input.upstreamWeightsText
      .split(/\r?\n/)
      .map((item) => item.trim())
      .filter(Boolean)
      .map((item) => Number(item))
      .filter((item) => Number.isFinite(item)),
    enabled: route?.enabled ?? true,
    enable_https: input.enableHttps,
    redirect_http: input.redirectHttp,
    limit_conn_per_server: route?.limit_conn_per_server ?? 0,
    limit_conn_per_ip: route?.limit_conn_per_ip ?? 0,
    limit_rate: input.limitRate,
    limit_req_per_ip: input.limitReqPerIP,
    cache_enabled: input.cacheEnabled,
    cache_policy: input.cachePolicy,
    cache_rules: input.cacheRulesText
      .split(/\r?\n/)
      .map((item) => item.trim())
      .filter(Boolean),
    custom_headers: [],
    basic_auth_enabled: route?.basic_auth_enabled ?? false,
    basic_auth_username: route?.basic_auth_username ?? '',
    basic_auth_password: '',
    upstream_type: input.upstreamType,
    pages_project_id: input.pagesProjectId
      ? Number(input.pagesProjectId)
      : null,
  };
}

export function useSiteConsole() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const queryClient = useQueryClient();
  const requestedRouteId = Number(searchParams.get('id'));
  const requestedDomainId = Number(searchParams.get('domain'));
  const [pendingRoute, setPendingRoute] = useState<ResourceRoute | null>(null);
  const [origin, setOrigin] = useState(emptyOrigin);
  const [siteName, setSiteName] = useState('');
  const [originId, setOriginId] = useState('');
  const [originScheme, setOriginScheme] = useState('http');
  const [originPort, setOriginPort] = useState('80');
  const [enableHttps, setEnableHttps] = useState(false);
  const [redirectHttp, setRedirectHttp] = useState(false);
  const [limitRate, setLimitRate] = useState('');
  const [limitReqPerIP, setLimitReqPerIP] = useState('');
  const [upstreamsText, setUpstreamsText] = useState('');
  const [upstreamWeightsText, setUpstreamWeightsText] = useState('');
  const [upstreamType, setUpstreamType] = useState('direct');
  const [pagesProjectId, setPagesProjectId] = useState('');
  const [cacheEnabled, setCacheEnabled] = useState(false);
  const [cachePolicy, setCachePolicy] = useState('static');
  const [cacheRulesText, setCacheRulesText] = useState('');
  const [childCertId, setChildCertId] = useState('');
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
  const certificates = useQuery({
    queryKey: ['custom', 'certificates'],
    queryFn: () => CustomService.listCertificates(),
  });
  const pages = useQuery({
    queryKey: ['custom', 'pages'],
    queryFn: () => CustomService.listPages(),
  });
  const policies = useQuery({
    queryKey: ['custom', 'policies'],
    queryFn: () => CustomService.listPolicies(),
  });
  const wafRules = useQuery({
    queryKey: ['custom', 'waf-rules'],
    queryFn: () => CustomService.listWafRules(),
  });

  const routeQuery = useQuery({
    queryKey: ['custom', 'route', requestedRouteId],
    queryFn: () => CustomService.getRoute(requestedRouteId),
    enabled: requestedRouteId > 0,
  });
  const routesQuery = useQuery({
    queryKey: ['custom', 'routes'],
    queryFn: () => CustomService.listRoutes(),
    enabled: requestedRouteId <= 0 && requestedDomainId > 0,
  });
  const route = useMemo(() => {
    if (pendingRoute) return pendingRoute;
    if (requestedRouteId > 0) return routeQuery.data;
    if (requestedDomainId > 0) {
      return (routesQuery.data ?? []).find((item) =>
        item.zone_domain_ids.includes(requestedDomainId),
      );
    }
    return undefined;
  }, [
    pendingRoute,
    requestedRouteId,
    routeQuery.data,
    requestedDomainId,
    routesQuery.data,
  ]);

  const zoneDetails = useQueries({
    queries: (zones.data ?? []).map((zone) => ({
      queryKey: ['custom', 'zone', zone.id],
      queryFn: () => CustomService.getZone(zone.id),
    })),
  });
  const domainsByZone = useMemo(() => {
    const map = new Map<number, ResourceDomain[]>();
    zoneDetails.forEach((query, index) => {
      const zone = (zones.data ?? [])[index];
      if (zone && query.data) {
        map.set(zone.id, query.data.domains);
      }
    });
    return map;
  }, [zoneDetails, zones.data]);

  const selectedDomain = useMemo<ResourceDomain | undefined>(() => {
    if (requestedDomainId > 0) {
      for (const domains of domainsByZone.values()) {
        const found = domains.find((item) => item.id === requestedDomainId);
        if (found) return found;
      }
    }
    const routeDomain = route?.zone_domains?.[0];
    if (!routeDomain) return undefined;
    // route.zone_domains 只包含精简字段，优先取域名分组里的完整记录以读取验证状态。
    for (const domains of domainsByZone.values()) {
      const found = domains.find((item) => item.id === routeDomain.id);
      if (found) return found;
    }
    return routeDomain;
  }, [requestedDomainId, domainsByZone, route]);

  const selectedZone = useMemo<ResourceZone | undefined>(() => {
    if (!selectedDomain) return undefined;
    return (zones.data ?? []).find(
      (zone) => zone.id === selectedDomain.zone_id,
    );
  }, [selectedDomain, zones.data]);

  const domainVerified = selectedDomain?.verification_status === 'verified';

  const routeWaf = useQuery({
    queryKey: ['custom', 'route-waf', route?.id],
    queryFn: () => CustomService.getRouteWaf(route!.id),
    enabled: Boolean(route),
  });

  useEffect(() => {
    if (!route) return;
    setSiteName(route.site_name ?? '');
    setOriginId(route.origin_id ? String(route.origin_id) : '');
    const originURL = route.origin_url ?? '';
    const urlMatch = originURL.match(/^(https?):\/\/[^/:]+(?::(\d+))?/);
    setOriginScheme(urlMatch?.[1] ?? 'http');
    setOriginPort(urlMatch?.[2] ?? '80');
    setEnableHttps(route.enable_https ?? false);
    setRedirectHttp(route.redirect_http ?? false);
    setLimitRate(route.limit_rate ?? '');
    setLimitReqPerIP(route.limit_req_per_ip ?? '');
    setUpstreamsText((route.upstream_list ?? []).slice(1).join('\n'));
    setUpstreamWeightsText(
      (route.upstream_weight_list ?? []).slice(1).join('\n'),
    );
    setUpstreamType(route.upstream_type ?? 'direct');
    setPagesProjectId(
      route.pages_project_id ? String(route.pages_project_id) : '',
    );
    setCacheEnabled(route.cache_enabled ?? false);
    setCachePolicy(route.cache_policy ?? 'static');
    setCacheRulesText((route.cache_rule_list ?? []).join('\n'));
  }, [route]);

  useEffect(() => {
    setChildCertId(
      selectedDomain?.cert_id ? String(selectedDomain.cert_id) : '',
    );
  }, [selectedDomain]);

  useEffect(() => {
    setWafDraft(routeWaf.data?.applied_ids ?? []);
  }, [routeWaf.data]);

  const refresh = useCallback(() => {
    void queryClient.invalidateQueries({ queryKey: ['custom'] });
  }, [queryClient]);

  const updateDomain = useMutation({
    mutationFn: ({
      domain,
      certId,
    }: {
      domain: ResourceDomain;
      certId: number | null;
    }) => {
      if (!selectedZone) throw new Error('请先选择域名分组');
      return CustomService.updateDomain(selectedZone.id, domain.id, {
        domain: domain.domain,
        cert_id: certId,
      });
    },
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
      if (!selectedDomain) throw new Error('请先选择域名');
      const payload = routePayload(route, {
        siteName,
        domainId: selectedDomain.id,
        originId,
        originScheme,
        originPort,
        enableHttps,
        redirectHttp,
        limitRate,
        limitReqPerIP,
        upstreamsText,
        upstreamWeightsText,
        upstreamType,
        pagesProjectId,
        cacheEnabled,
        cachePolicy,
        cacheRulesText,
      });
      return route
        ? CustomService.updateRoute(route.id, payload)
        : CustomService.createRoute(payload);
    },
    onSuccess: (created) => {
      if (!route) {
        setPendingRoute(created);
        router.replace(`/resources/detail?id=${created.id}`);
      }
      refresh();
      toast.success('域名配置已保存');
    },
    onError: (error) => toast.error(errorMessage(error, '保存域名配置失败')),
  });

  const deleteOrigin = useMutation({
    mutationFn: (id: number) => CustomService.deleteOrigin(id),
    onSuccess: () => {
      refresh();
      toast.success('源站已删除');
    },
    onError: (error) => toast.error(errorMessage(error, '删除源站失败')),
  });

  const publish = useMutation({
    mutationFn: () => CustomService.publish(),
    onSuccess: (data) => toast.success(`已部署版本 ${data.version}`),
    onError: (error) => toast.error(errorMessage(error, '部署失败')),
  });

  const verifyDomain = useMutation({
    mutationFn: (domainId: number) => CustomService.verifySite(domainId),
    onSuccess: () => {
      refresh();
      toast.success('域名验证成功');
    },
    onError: (error) => toast.error(errorMessage(error, '域名验证失败')),
  });

  const createWaf = useMutation({
    mutationFn: () => {
      if (!selectedDomain) throw new Error('请先选择域名');
      return CustomService.createWafRule({
        name: wafName,
        host: selectedDomain.domain,
      });
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
      if (!route) throw new Error('请先保存域名配置');
      return CustomService.updateRouteWaf(route.id, wafDraft);
    },
    onSuccess: () => {
      void routeWaf.refetch();
      toast.success('防火墙绑定已保存');
    },
    onError: (error) => toast.error(errorMessage(error, '保存防火墙绑定失败')),
  });

  const deleteZone = useMutation({
    mutationFn: () => {
      if (!selectedZone) throw new Error('找不到域名分组');
      return CustomService.deleteZone(selectedZone.id);
    },
    onSuccess: () => {
      toast.success('站点已删除');
      router.push('/resources');
    },
    onError: (error) => toast.error(errorMessage(error, '删除失败')),
  });

  return {
    route,
    selectedDomain,
    selectedZone,
    domainVerified,
    domainsByZone,
    zones: zones.data ?? [],
    origins: origins.data ?? [],
    certificates: certificates.data ?? [],
    pages: pages.data ?? [],
    policies: policies.data,
    wafRules: wafRules.data ?? [],
    routeWaf,
    routeLoading: routeQuery.isLoading,
    loading:
      zones.isLoading ||
      origins.isLoading ||
      certificates.isLoading ||
      pages.isLoading ||
      policies.isLoading ||
      wafRules.isLoading ||
      routeQuery.isLoading ||
      routesQuery.isLoading ||
      zoneDetails.some((query) => query.isLoading),
    origin,
    setOrigin,
    siteName,
    setSiteName,
    originId,
    setOriginId,
    originScheme,
    setOriginScheme,
    originPort,
    setOriginPort,
    enableHttps,
    setEnableHttps,
    redirectHttp,
    setRedirectHttp,
    limitRate,
    setLimitRate,
    limitReqPerIP,
    setLimitReqPerIP,
    upstreamsText,
    setUpstreamsText,
    upstreamWeightsText,
    setUpstreamWeightsText,
    upstreamType,
    setUpstreamType,
    pagesProjectId,
    setPagesProjectId,
    cacheEnabled,
    setCacheEnabled,
    cachePolicy,
    setCachePolicy,
    cacheRulesText,
    setCacheRulesText,
    childCertId,
    setChildCertId,
    wafName,
    setWafName,
    wafDraft,
    setWafDraft,
    refresh,
    updateDomain,
    createOrigin,
    deleteOrigin,
    saveRoute,
    publish,
    verifyDomain,
    createWaf,
    saveWaf,
    deleteZone,
  };
}
