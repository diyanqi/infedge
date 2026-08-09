// 站点控制台：安全防护 Tab（限流 + WAF）。
'use client';

import { Checkbox } from '@/components/ui/checkbox';
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
import { Plus, Settings2, ShieldCheck, ShieldAlert } from 'lucide-react';

import { Badge } from '@/components/ui/badge';

import type { useSiteConsole } from '../use-site-console';

export function SecurityTab({
  console: value,
}: {
  console: ReturnType<typeof useSiteConsole>;
}) {
  const {
    route,
    selectedDomain,
    domainVerified,
    limitRate,
    setLimitRate,
    limitReqPerIP,
    setLimitReqPerIP,
    wafName,
    setWafName,
    wafDraft,
    setWafDraft,
    wafRules,
    policies,
    createWaf,
    saveWaf,
  } = value;

  const customRules = (wafRules ?? []).filter((rule) => !rule.is_global);
  const globalRules = policies?.global_rules ?? [];
  const verified = domainVerified;

  const toggleRule = (id: number) => {
    setWafDraft((draft) =>
      draft.includes(id) ? draft.filter((item) => item !== id) : [...draft, id],
    );
  };

  return (
    <div className='space-y-5'>
      <Card>
        <CardHeader>
          <CardTitle className='text-base'>限流策略</CardTitle>
          <CardDescription>留空则继承全局默认策略。</CardDescription>
        </CardHeader>
        <CardContent className='grid gap-3 md:grid-cols-2'>
          <div className='flex flex-col gap-2'>
            <Label htmlFor='limit-rate'>单连接限速</Label>
            <Input
              id='limit-rate'
              value={limitRate}
              onChange={(event) => setLimitRate(event.target.value)}
              placeholder='留空使用全局默认'
            />
          </div>
          <div className='flex flex-col gap-2'>
            <Label htmlFor='limit-req'>单 IP 请求限流</Label>
            <Input
              id='limit-req'
              value={limitReqPerIP}
              onChange={(event) => setLimitReqPerIP(event.target.value)}
              placeholder='留空使用全局默认'
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className='flex flex-wrap items-start justify-between gap-3'>
            <div className='flex flex-col gap-1'>
              <CardTitle className='text-base'>防火墙</CardTitle>
              <CardDescription>
                规则仅绑定当前域名；全局规则始终优先且只读。
              </CardDescription>
            </div>
            <ShieldCheck className='size-5 text-primary' />
          </div>
        </CardHeader>
        <CardContent className='flex flex-col gap-5'>
          <div className='flex flex-col gap-2'>
            <Label htmlFor='waf-name'>创建当前域名规则</Label>
            <div className='flex gap-2'>
              <Input
                id='waf-name'
                value={wafName}
                onChange={(event) => setWafName(event.target.value)}
                placeholder={
                  selectedDomain
                    ? `例如：保护 ${selectedDomain.domain}`
                    : '规则名称'
                }
              />
              <Button
                onClick={() => createWaf.mutate()}
                disabled={!verified || !wafName || createWaf.isPending}
              >
                <Plus data-icon='inline-start' />
                创建规则
              </Button>
            </div>
          </div>
          {route ? (
            <div className='flex flex-col gap-3'>
              <Label>绑定自定义规则</Label>
              {customRules.map((rule) => (
                <label
                  key={rule.id}
                  className='flex cursor-pointer items-center gap-2 rounded-md border px-3 py-2 text-sm'
                >
                  <Checkbox
                    checked={wafDraft.includes(rule.id)}
                    onCheckedChange={() => toggleRule(rule.id)}
                  />
                  <span className='min-w-0 flex-1 truncate'>{rule.name}</span>
                  <Badge variant='outline'>{rule.host}</Badge>
                </label>
              ))}
              {!customRules.length && (
                <p className='text-sm text-muted-foreground'>
                  暂无自定义规则。
                </p>
              )}
              <Button
                variant='outline'
                className='w-fit'
                onClick={() => saveWaf.mutate()}
                disabled={!verified || saveWaf.isPending}
              >
                保存防火墙绑定
              </Button>
            </div>
          ) : (
            <p className='text-sm text-muted-foreground'>
              请先在“源站”页保存站点配置，再绑定防火墙规则。
            </p>
          )}
          <div className='flex flex-col gap-2 border-t pt-4'>
            <Label>全局规则（只读）</Label>
            {globalRules.map((rule) => (
              <div key={rule.id} className='flex items-center gap-2 text-sm'>
                <ShieldAlert className='size-4 text-primary' />
                <span>{rule.name}</span>
                <Badge variant='secondary'>全局优先</Badge>
              </div>
            ))}
            {!globalRules.length && (
              <span className='text-sm text-muted-foreground'>
                暂无全局规则
              </span>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
