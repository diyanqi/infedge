// 站点控制台：HTTPS / TLS 证书 Tab。
'use client';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';

import { certificateCoversDomain } from '../use-site-console';
import type { useSiteConsole } from '../use-site-console';

export function HttpsTab({
  console: value,
}: {
  console: ReturnType<typeof useSiteConsole>;
}) {
  const {
    selectedDomain,
    certificates,
    enableHttps,
    setEnableHttps,
    redirectHttp,
    setRedirectHttp,
    childCertId,
    setChildCertId,
    updateDomain,
  } = value;

  const domain = selectedDomain?.domain ?? '';

  return (
    <div className='space-y-5'>
      <Card>
        <CardHeader>
          <CardTitle className='text-base'>HTTPS 与 TLS 证书</CardTitle>
          <CardDescription>
            开启 HTTPS 后自动为当前域名选择可用证书。
          </CardDescription>
        </CardHeader>
        <CardContent className='flex flex-col gap-5'>
          <div className='flex items-center gap-3'>
            <Switch
              checked={enableHttps}
              onCheckedChange={(enabled) => {
                setEnableHttps(enabled);
                if (!enabled || !selectedDomain) return;
                if (selectedDomain.cert_id) return;
                const certificate = (certificates ?? []).find((item) =>
                  certificateCoversDomain(item, selectedDomain.domain),
                );
                if (certificate) {
                  setChildCertId(String(certificate.id));
                  updateDomain.mutate({
                    domain: selectedDomain,
                    certId: certificate.id,
                  });
                }
              }}
              aria-label='启用 HTTPS'
            />
            <span className='text-sm'>
              {enableHttps ? 'HTTPS 已启用' : 'HTTPS 未启用'}
            </span>
            <Badge variant='outline'>
              {selectedDomain?.cert_id ? '已绑定证书' : '未绑定证书'}
            </Badge>
          </div>

          <div className='flex flex-col gap-2'>
            <Label>TLS 证书</Label>
            <Select
              value={childCertId}
              onValueChange={setChildCertId}
              disabled={!domain}
            >
              <SelectTrigger>
                <SelectValue placeholder='选择证书（留空则使用平台默认）' />
              </SelectTrigger>
              <SelectContent>
                {(certificates ?? []).map((item) => (
                  <SelectItem key={item.id} value={String(item.id)}>
                    {item.name}
                    {item.primary_domain ? ` · ${item.primary_domain}` : ''}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              className='w-fit'
              size='sm'
              variant='outline'
              disabled={
                !domain ||
                !selectedDomain ||
                updateDomain.isPending ||
                !childCertId
              }
              onClick={() =>
                selectedDomain &&
                updateDomain.mutate({
                  domain: selectedDomain,
                  certId: childCertId ? Number(childCertId) : null,
                })
              }
            >
              保存证书
            </Button>
          </div>

          <div className='flex items-center gap-3 border-t pt-4'>
            <Switch
              checked={redirectHttp}
              onCheckedChange={setRedirectHttp}
              disabled={!enableHttps}
              aria-label='HTTP 跳转 HTTPS'
            />
            <span className='text-sm'>HTTP 自动跳转 HTTPS</span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
