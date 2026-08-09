// 站点控制台：域名 Tab。
'use client';

import { CheckCircle2, Link2, ShieldAlert } from 'lucide-react';

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

import type { useSiteConsole } from '../use-site-console';

export function DomainsTab({
  console: value,
}: {
  console: ReturnType<typeof useSiteConsole>;
}) {
  const { selectedZone, selectedDomain, domainsByZone, verifyDomain } = value;
  const domains = selectedZone
    ? (domainsByZone.get(selectedZone.id) ??
      (selectedDomain ? [selectedDomain] : []))
    : selectedDomain
      ? [selectedDomain]
      : [];

  return (
    <div className='space-y-5'>
      <Alert>
        <Link2 />
        <AlertTitle>DNS 接入要求</AlertTitle>
        <AlertDescription>
          所有接入域名必须将 CNAME 指向 <strong>cname.edge.infvar.com</strong>{' '}
          后再部署。禁止自行优选 IP，发现后将封号且不可申诉。
        </AlertDescription>
      </Alert>

      <Card>
        <CardHeader>
          <CardTitle className='text-base'>域名列表</CardTitle>
          <CardDescription>
            该域名分组下的全部域名；未验证域名需完成 DNS TXT 所有权验证。
          </CardDescription>
        </CardHeader>
        <CardContent className='flex flex-col gap-3'>
          {domains.length === 0 && (
            <p className='py-4 text-center text-sm text-muted-foreground'>
              暂无域名，请先在接入页添加域名。
            </p>
          )}
          {domains.map((domain) => {
            const verified = domain.verification_status === 'verified';
            return (
              <div key={domain.id} className='rounded-md border px-4 py-3'>
                <div className='flex flex-wrap items-center justify-between gap-3'>
                  <div className='flex min-w-0 items-center gap-2'>
                    {verified ? (
                      <CheckCircle2 className='size-4 shrink-0 text-primary' />
                    ) : (
                      <ShieldAlert className='size-4 shrink-0 text-muted-foreground' />
                    )}
                    <p className='truncate font-mono text-sm font-medium'>
                      {domain.domain}
                    </p>
                    <Badge variant={verified ? 'secondary' : 'outline'}>
                      {verified ? 'TXT 已验证' : '待 TXT 验证'}
                    </Badge>
                  </div>
                  {!verified && (
                    <Button
                      size='sm'
                      variant='outline'
                      disabled={verifyDomain.isPending}
                      onClick={() => verifyDomain.mutate(domain.id)}
                    >
                      验证域名
                    </Button>
                  )}
                </div>
                {!verified && domain.verification_token ? (
                  <div className='mt-3 rounded-md bg-muted p-3 text-xs'>
                    <p className='mb-1 text-muted-foreground'>
                      在 DNS 服务商创建 TXT 记录，名称为{' '}
                      <code className='rounded bg-background px-1 py-0.5'>
                        _openflare-verification.{domain.domain}
                      </code>
                      ，值为：
                    </p>
                    <code className='block break-all rounded bg-background px-2 py-1'>
                      {domain.verification_token}
                    </code>
                  </div>
                ) : null}
              </div>
            );
          })}
        </CardContent>
      </Card>
    </div>
  );
}
