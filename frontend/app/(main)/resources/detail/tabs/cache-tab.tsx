// 站点控制台：缓存 Tab。
'use client';

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
import { Textarea } from '@/components/ui/textarea';

import type { useSiteConsole } from '../use-site-console';

export function CacheTab({
  console: value,
}: {
  console: ReturnType<typeof useSiteConsole>;
}) {
  const {
    cacheEnabled,
    setCacheEnabled,
    cachePolicy,
    setCachePolicy,
    cacheRulesText,
    setCacheRulesText,
  } = value;

  return (
    <Card>
      <CardHeader>
        <CardTitle className='text-base'>边缘缓存</CardTitle>
        <CardDescription>
          开启后按策略缓存边缘响应，降低源站压力。
        </CardDescription>
      </CardHeader>
      <CardContent className='flex flex-col gap-5'>
        <div className='flex items-center gap-3'>
          <Switch
            checked={cacheEnabled}
            onCheckedChange={setCacheEnabled}
            aria-label='启用边缘缓存'
          />
          <Label>边缘缓存</Label>
          {cacheEnabled ? (
            <Select value={cachePolicy} onValueChange={setCachePolicy}>
              <SelectTrigger className='w-44'>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value='static'>标准静态资源</SelectItem>
                <SelectItem value='all'>所有可缓存 GET</SelectItem>
                <SelectItem value='suffix'>指定后缀</SelectItem>
                <SelectItem value='path_prefix'>路径前缀</SelectItem>
                <SelectItem value='path_exact'>精确路径</SelectItem>
              </SelectContent>
            </Select>
          ) : null}
        </div>
        {cacheEnabled && cachePolicy !== 'static' && cachePolicy !== 'all' ? (
          <div className='flex flex-col gap-2'>
            <Label htmlFor='cache-rules'>缓存规则</Label>
            <Textarea
              id='cache-rules'
              value={cacheRulesText}
              onChange={(event) => setCacheRulesText(event.target.value)}
              placeholder='每行一条缓存规则'
              rows={4}
            />
          </div>
        ) : null}
      </CardContent>
    </Card>
  );
}
