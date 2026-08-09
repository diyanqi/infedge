// 普通用户域名接入页：添加域名、完成 TXT 所有权验证后进入站点控制台。
'use client';

import Link from 'next/link';
import { useEffect, useMemo, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import {
  useMutation,
  useQueries,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import {
  ArrowLeft,
  ArrowRight,
  CheckCircle2,
  Globe,
  Link2,
  Plus,
  ShieldAlert,
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
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { CustomService } from '@/lib/services/custom';
import type { ResourceDomain } from '@/lib/services/custom';

function errorMessage(error: unknown, fallback: string) {
  return error instanceof Error ? error.message : fallback;
}

export default function ConfigurePage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const requestedDomainId = Number(searchParams.get('domain'));
  const queryClient = useQueryClient();
  const [selectedZoneId, setSelectedZoneId] = useState<number | null>(null);
  const [selectedDomainId, setSelectedDomainId] = useState<number | null>(null);
  const [rootDomain, setRootDomain] = useState('');

  const zones = useQuery({
    queryKey: ['custom', 'zones'],
    queryFn: () => CustomService.listZones(),
  });
  const domainQueries = useQueries({
    queries: (zones.data ?? []).map((zone) => ({
      queryKey: ['custom', 'zone', zone.id],
      queryFn: () => CustomService.getZone(zone.id),
    })),
  });
  const domainsByZone = useMemo(() => {
    const map = new Map<number, ResourceDomain[]>();
    domainQueries.forEach((query, index) => {
      const zone = (zones.data ?? [])[index];
      if (zone && query.data) {
        map.set(zone.id, query.data.domains);
      }
    });
    return map;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [domainQueries, zones.data]);

  const selectedZone = useMemo(
    () => (zones.data ?? []).find((zone) => zone.id === selectedZoneId) ?? null,
    [selectedZoneId, zones.data],
  );
  const zoneDomains = selectedZone
    ? (domainsByZone.get(selectedZone.id) ?? [])
    : [];

  useEffect(() => {
    if (requestedDomainId > 0) {
      for (const [zoneId, domains] of domainsByZone.entries()) {
        const found = domains.find((domain) => domain.id === requestedDomainId);
        if (found) {
          setSelectedZoneId(zoneId);
          setSelectedDomainId(found.id);
          return;
        }
      }
    }
  }, [requestedDomainId, domainsByZone]);

  const selectedDomain = useMemo(() => {
    if (requestedDomainId > 0) {
      for (const domains of domainsByZone.values()) {
        const found = domains.find((domain) => domain.id === requestedDomainId);
        if (found) return found;
      }
    }
    if (!selectedZone) return null;
    return (
      zoneDomains.find((domain) => domain.id === selectedDomainId) ??
      zoneDomains[0] ??
      null
    );
  }, [
    requestedDomainId,
    domainsByZone,
    selectedZone,
    zoneDomains,
    selectedDomainId,
  ]);

  const refresh = () => {
    void queryClient.invalidateQueries({ queryKey: ['custom'] });
  };

  const createZone = useMutation({
    mutationFn: () => CustomService.createSite(rootDomain),
    onSuccess: (site) => {
      setRootDomain('');
      refresh();
      router.replace(`/resources/configure?domain=${site.domain.id}`);
      toast.success('域名已添加，请按提示完成 DNS TXT 验证');
    },
    onError: (error) => toast.error(errorMessage(error, '添加域名失败')),
  });

  const verifyDomain = useMutation({
    mutationFn: (domainId: number) => CustomService.verifySite(domainId),
    onSuccess: () => {
      refresh();
      toast.success('域名验证成功');
    },
    onError: (error) => toast.error(errorMessage(error, '域名验证失败')),
  });

  const removeZone = useMutation({
    mutationFn: (zoneId: number) => CustomService.deleteZone(zoneId),
    onSuccess: () => {
      setSelectedZoneId(null);
      setSelectedDomainId(null);
      refresh();
      toast.success('域名分组已删除');
    },
    onError: (error) => toast.error(errorMessage(error, '删除失败')),
  });

  return (
    <div className='w-full flex flex-col gap-6 py-6 px-1'>
      <div className='flex items-center gap-2'>
        <Button variant='ghost' size='icon' asChild>
          <Link href='/resources' aria-label='返回网站列表'>
            <ArrowLeft />
          </Link>
        </Button>
        <Globe className='size-5 text-primary' />
        <h1 className='text-2xl font-semibold tracking-tight'>网站接入</h1>
      </div>

      <Alert variant='destructive'>
        <ShieldAlert />
        <AlertTitle>域名接入要求</AlertTitle>
        <AlertDescription>
          所有接入域名必须将 CNAME 指向 <strong>cname.edge.infvar.com</strong>
          。禁止自行优选 IP，发现后将封号且不可申诉；该 CNAME
          已每小时进行全国拨测优选。
        </AlertDescription>
      </Alert>

      <Card>
        <CardHeader>
          <CardTitle className='text-base'>添加 CDN 域名</CardTitle>
          <CardDescription>
            输入根域或子域即可创建站点，系统会自动完成后台归组。
          </CardDescription>
        </CardHeader>
        <CardContent className='flex flex-col gap-3 lg:flex-row lg:items-end'>
          <div className='flex-1 flex flex-col gap-2'>
            <Label htmlFor='root-domain'>域名</Label>
            <Input
              id='root-domain'
              value={rootDomain}
              onChange={(event) => setRootDomain(event.target.value)}
              placeholder='example.com 或 www.example.com'
            />
          </div>
          <Button
            onClick={() => createZone.mutate()}
            disabled={!rootDomain || createZone.isPending}
          >
            <Plus data-icon='inline-start' />
            添加域名
          </Button>
        </CardContent>
      </Card>

      <div className='grid gap-6 lg:grid-cols-[220px_minmax(0,1fr)]'>
        <Card className='h-fit'>
          <CardHeader className='pb-3'>
            <CardTitle className='text-base'>域名分组</CardTitle>
          </CardHeader>
          <CardContent className='flex flex-col gap-1'>
            {(zones.data ?? []).map((zone) => (
              <Button
                key={zone.id}
                variant={zone.id === selectedZone?.id ? 'secondary' : 'ghost'}
                className='justify-start'
                onClick={() => {
                  setSelectedZoneId(zone.id);
                  setSelectedDomainId(
                    (domainsByZone.get(zone.id) ?? [])[0]?.id ?? null,
                  );
                }}
              >
                <Globe data-icon='inline-start' />
                <span className='truncate'>{zone.domain}</span>
              </Button>
            ))}
            {!zones.data?.length && (
              <p className='px-2 py-4 text-sm text-muted-foreground'>
                暂无域名
              </p>
            )}
          </CardContent>
        </Card>

        <div className='flex min-w-0 flex-col gap-6'>
          {selectedZone ? (
            <Card>
              <CardHeader className='flex flex-row items-start justify-between gap-3'>
                <div className='flex flex-col gap-1'>
                  <CardTitle className='text-base'>
                    {selectedZone.domain}
                  </CardTitle>
                  <CardDescription>
                    先选择域名，完成验证后进入站点控制台配置源站、HTTPS、缓存和
                    WAF。
                  </CardDescription>
                </div>
                <Button
                  variant='ghost'
                  size='icon'
                  onClick={() => removeZone.mutate(selectedZone.id)}
                  disabled={removeZone.isPending}
                  aria-label='删除域名分组'
                >
                  <Trash2 />
                </Button>
              </CardHeader>
              <CardContent className='flex flex-col gap-4'>
                <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-3'>
                  {zoneDomains.map((domain) => (
                    <Button
                      key={domain.id}
                      variant={
                        domain.id === selectedDomain?.id
                          ? 'secondary'
                          : 'outline'
                      }
                      className='h-auto min-h-16 justify-between px-4 py-3 text-left'
                      onClick={() => setSelectedDomainId(domain.id)}
                    >
                      <span className='flex min-w-0 flex-col gap-1'>
                        <span className='truncate'>{domain.domain}</span>
                        <span className='text-xs text-muted-foreground'>
                          {domain.proxy_route_id
                            ? '已配置站点'
                            : '尚未配置站点'}
                        </span>
                      </span>
                      {domain.verification_status === 'verified' ? (
                        <CheckCircle2 />
                      ) : (
                        <ShieldAlert />
                      )}
                    </Button>
                  ))}
                </div>
                {selectedDomain ? (
                  <div className='flex flex-col gap-3 border-t pt-4'>
                    <div className='flex flex-wrap items-center justify-between gap-3'>
                      <div className='flex items-center gap-2'>
                        <Globe className='size-4 text-primary' />
                        <span className='font-medium'>
                          {selectedDomain.domain}
                        </span>
                        <Badge variant='outline'>
                          {selectedDomain.verification_status === 'verified'
                            ? 'TXT 已验证'
                            : '待 TXT 验证'}
                        </Badge>
                      </div>
                      {selectedDomain.verification_status !== 'verified' && (
                        <Button
                          size='sm'
                          variant='outline'
                          onClick={() => verifyDomain.mutate(selectedDomain.id)}
                          disabled={verifyDomain.isPending}
                        >
                          验证域名
                        </Button>
                      )}
                    </div>
                    <Alert>
                      <Link2 />
                      <AlertTitle>DNS 配置</AlertTitle>
                      <AlertDescription>
                        请将 {selectedDomain.domain} 的 CNAME 指向{' '}
                        <strong>cname.edge.infvar.com</strong> 后再部署。
                      </AlertDescription>
                    </Alert>
                    {selectedDomain.verification_status !== 'verified' &&
                    selectedDomain.verification_token ? (
                      <Alert>
                        <ShieldAlert />
                        <AlertTitle>完成域名所有权验证</AlertTitle>
                        <AlertDescription className='space-y-2'>
                          <p>
                            请在 DNS 服务商创建 TXT 记录，名称为{' '}
                            <code>
                              _openflare-verification.{selectedDomain.domain}
                            </code>
                            ， 值为：
                          </p>
                          <code className='block break-all rounded bg-muted px-2 py-1 text-xs'>
                            {selectedDomain.verification_token}
                          </code>
                        </AlertDescription>
                      </Alert>
                    ) : null}
                    <Button asChild className='w-fit'>
                      <Link
                        href={`/resources/detail?domain=${selectedDomain.id}`}
                      >
                        进入站点控制台
                        <ArrowRight data-icon='inline-end' />
                      </Link>
                    </Button>
                  </div>
                ) : (
                  <p className='py-4 text-center text-sm text-muted-foreground'>
                    该分组下暂无域名。
                  </p>
                )}
              </CardContent>
            </Card>
          ) : (
            <Card>
              <CardContent className='py-12 text-center text-sm text-muted-foreground'>
                添加域名后，从左侧选择一个域名分组。
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
