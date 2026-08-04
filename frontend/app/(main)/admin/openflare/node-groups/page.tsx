'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Boxes, Pencil, Plus, Trash2 } from 'lucide-react';
import { toast } from 'sonner';
import { CustomService } from '@/lib/services/custom';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

export default function NodeGroupsPage() {
  const qc = useQueryClient();
  const [editing, setEditing] = useState<number | null>(null);
  const [name, setName] = useState('');
  const [limit, setLimit] = useState('0');
  const [nodeIDs, setNodeIDs] = useState('');
  const groups = useQuery({
    queryKey: ['custom', 'node-groups'],
    queryFn: () => CustomService.listNodeGroups(),
  });
  const payload = () => ({
    name,
    monthly_bytes_limit: Number(limit),
    node_ids: nodeIDs
      .split(',')
      .map((value) => Number(value.trim()))
      .filter((value) => value > 0),
  });
  const save = useMutation({
    mutationFn: () =>
      editing
        ? CustomService.updateNodeGroup(editing, payload())
        : CustomService.createNodeGroup(payload()),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ['custom', 'node-groups'] });
      setEditing(null);
      setName('');
      setLimit('0');
      setNodeIDs('');
      toast.success('节点组已保存');
    },
    onError: (e) => toast.error(e instanceof Error ? e.message : '保存失败'),
  });
  const remove = useMutation({
    mutationFn: (id: number) => CustomService.deleteNodeGroup(id),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ['custom', 'node-groups'] }),
  });

  return (
    <div className='w-full space-y-6 py-6 px-1'>
      <div className='flex items-center gap-2'>
        <Boxes className='size-5 text-primary' />
        <h1 className='text-2xl font-semibold tracking-tight'>节点组流量</h1>
      </div>
      <Card>
        <CardHeader>
          <CardTitle className='text-base'>
            {editing ? '编辑节点组' : '新增节点组'}
          </CardTitle>
        </CardHeader>
        <CardContent className='grid gap-3 md:grid-cols-4'>
          <div>
            <Label>名称</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div>
            <Label>月总流量（字节）</Label>
            <Input
              type='number'
              min='0'
              value={limit}
              onChange={(e) => setLimit(e.target.value)}
            />
          </div>
          <div>
            <Label>节点 ID（逗号分隔）</Label>
            <Input
              value={nodeIDs}
              onChange={(e) => setNodeIDs(e.target.value)}
              placeholder='1,2,3'
            />
          </div>
          <div className='flex items-end gap-2'>
            <Button
              onClick={() => save.mutate()}
              disabled={!name || save.isPending}
            >
              <Plus className='mr-2 size-4' />
              保存
            </Button>
            {editing && (
              <Button
                variant='outline'
                onClick={() => {
                  setEditing(null);
                  setName('');
                  setLimit('0');
                  setNodeIDs('');
                }}
              >
                取消
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
      <div className='grid gap-3 md:grid-cols-2'>
        {(groups.data ?? []).map((group) => (
          <Card key={group.id} className='border-dashed'>
            <CardContent className='flex items-center justify-between gap-3 pt-6'>
              <div>
                <p className='font-medium'>{group.name}</p>
                <p className='text-sm text-muted-foreground'>
                  月总量 {group.monthly_bytes_limit} 字节 ·{' '}
                  {group.nodes?.length ?? 0} 个节点
                </p>
                <p className='text-xs text-muted-foreground'>
                  {group.nodes
                    ?.map((node) => node.name || node.node_id)
                    .join(', ')}
                </p>
              </div>
              <div className='flex gap-2'>
                <Button
                  variant='outline'
                  size='sm'
                  onClick={() => {
                    setEditing(group.id);
                    setName(group.name);
                    setLimit(String(group.monthly_bytes_limit));
                    setNodeIDs(
                      (group.nodes ?? [])
                        .map((node) => String(node.id))
                        .join(','),
                    );
                  }}
                >
                  <Pencil className='mr-1 size-3' />
                  编辑
                </Button>
                <Button
                  variant='ghost'
                  size='icon'
                  onClick={() => remove.mutate(group.id)}
                  aria-label='删除节点组'
                >
                  <Trash2 className='size-4' />
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
