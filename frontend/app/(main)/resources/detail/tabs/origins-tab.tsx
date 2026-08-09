// 站点控制台：源站 Tab。
'use client';

import { Plus, Server, Trash2 } from 'lucide-react';

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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';

import type { useSiteConsole } from '../use-site-console';

export function OriginsTab({
  console: value,
}: {
  console: ReturnType<typeof useSiteConsole>;
}) {
  const {
    origins,
    pages,
    origin,
    setOrigin,
    originId,
    setOriginId,
    originScheme,
    setOriginScheme,
    originPort,
    setOriginPort,
    upstreamType,
    setUpstreamType,
    pagesProjectId,
    setPagesProjectId,
    upstreamsText,
    setUpstreamsText,
    upstreamWeightsText,
    setUpstreamWeightsText,
    createOrigin,
  } = value;

  return (
    <div className='space-y-5'>
      <Card>
        <CardHeader>
          <CardTitle className='text-base'>源站配置</CardTitle>
          <CardDescription>
            选择源站并配置协议、端口与备用源站。
          </CardDescription>
        </CardHeader>
        <CardContent className='flex flex-col gap-5'>
          <div className='grid gap-4 md:grid-cols-2'>
            <div className='flex flex-col gap-2'>
              <Label>源站</Label>
              <Select value={originId} onValueChange={setOriginId}>
                <SelectTrigger>
                  <SelectValue placeholder='选择源站' />
                </SelectTrigger>
                <SelectContent>
                  {(origins ?? []).map((item) => (
                    <SelectItem key={item.id} value={String(item.id)}>
                      {item.name} · {item.address}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <div className='grid grid-cols-2 gap-2'>
                <Select value={originScheme} onValueChange={setOriginScheme}>
                  <SelectTrigger aria-label='源站协议'>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value='http'>HTTP</SelectItem>
                    <SelectItem value='https'>HTTPS</SelectItem>
                  </SelectContent>
                </Select>
                <Input
                  value={originPort}
                  onChange={(event) => setOriginPort(event.target.value)}
                  placeholder='端口'
                  aria-label='源站端口'
                />
              </div>
            </div>
            <div className='flex flex-col gap-2'>
              <Label>源站类型</Label>
              <Select value={upstreamType} onValueChange={setUpstreamType}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='direct'>直连源站</SelectItem>
                  <SelectItem value='pages'>Pages 项目</SelectItem>
                </SelectContent>
              </Select>
              {upstreamType === 'pages' ? (
                <Select
                  value={pagesProjectId}
                  onValueChange={setPagesProjectId}
                >
                  <SelectTrigger>
                    <SelectValue placeholder='选择 Pages 项目' />
                  </SelectTrigger>
                  <SelectContent>
                    {(pages ?? []).map((project) => (
                      <SelectItem key={project.id} value={String(project.id)}>
                        {project.name} · {project.slug}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : null}
            </div>
          </div>
          {upstreamType === 'direct' ? (
            <div className='grid gap-4 md:grid-cols-2'>
              <div className='flex flex-col gap-2'>
                <Label htmlFor='upstreams'>备用源站地址</Label>
                <Textarea
                  id='upstreams'
                  value={upstreamsText}
                  onChange={(event) => setUpstreamsText(event.target.value)}
                  placeholder='每行一个 http(s)://host:port 地址，第一行是主源站'
                  rows={4}
                />
              </div>
              <div className='flex flex-col gap-2'>
                <Label htmlFor='upstream-weights'>源站权重</Label>
                <Textarea
                  id='upstream-weights'
                  value={upstreamWeightsText}
                  onChange={(event) =>
                    setUpstreamWeightsText(event.target.value)
                  }
                  placeholder='每行一个权重，第一行是主源站，例如 3、1'
                  rows={4}
                />
              </div>
            </div>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className='text-base'>源站资源</CardTitle>
          <CardDescription>源站可被自己的多个域名配置复用。</CardDescription>
        </CardHeader>
        <CardContent className='flex flex-col gap-4'>
          <div className='grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)_auto]'>
            <Input
              placeholder='名称'
              value={origin.name}
              onChange={(event) =>
                setOrigin({ ...origin, name: event.target.value })
              }
            />
            <Input
              placeholder='源站地址，如 10.0.0.8'
              value={origin.address}
              onChange={(event) =>
                setOrigin({ ...origin, address: event.target.value })
              }
            />
            <Input
              placeholder='备注'
              value={origin.remark}
              onChange={(event) =>
                setOrigin({ ...origin, remark: event.target.value })
              }
            />
            <Button
              onClick={() => createOrigin.mutate()}
              disabled={!origin.address || createOrigin.isPending}
            >
              <Plus data-icon='inline-start' />
              创建
            </Button>
          </div>
          <div className='grid gap-3 md:grid-cols-2'>
            {(origins ?? []).map((item) => (
              <div
                key={item.id}
                className='flex items-center justify-between gap-3 rounded-md border px-3 py-3'
              >
                <div className='flex min-w-0 items-center gap-2'>
                  <Server className='size-4 shrink-0 text-primary' />
                  <div className='min-w-0'>
                    <p className='truncate text-sm font-medium'>{item.name}</p>
                    <p className='truncate text-xs text-muted-foreground'>
                      {item.address}
                    </p>
                  </div>
                </div>
                <Button
                  variant='ghost'
                  size='icon'
                  onClick={() => value.deleteOrigin.mutate(item.id)}
                  aria-label='删除源站'
                >
                  <Trash2 />
                </Button>
              </div>
            ))}
            {!origins?.length && (
              <p className='col-span-full py-4 text-center text-sm text-muted-foreground'>
                暂无源站，创建后即可在左侧选择使用。
              </p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
