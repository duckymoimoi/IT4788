---
name: antd-react-patterns  
description: Common Ant Design + React patterns for admin dashboards. Use when building tables, forms, modals, charts, or canvas-based visualizations in the Hospital Admin Panel.
---

# Ant Design + React Patterns

## 1. CRUD Table with Actions

```jsx
import { Table, Button, Space, Popconfirm, message } from 'antd';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

export default function CrudTable({ queryKey, fetchFn, deleteFn, columns }) {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey, queryFn: fetchFn });
  
  const deleteMutation = useMutation({
    mutationFn: deleteFn,
    onSuccess: () => {
      message.success('Deleted');
      queryClient.invalidateQueries({ queryKey });
    },
  });

  const actionColumn = {
    title: 'Actions',
    render: (_, record) => (
      <Space>
        <Button type="link" onClick={() => onEdit(record)}>Edit</Button>
        <Popconfirm title="Delete?" onConfirm={() => deleteMutation.mutate(record.id)}>
          <Button type="link" danger>Delete</Button>
        </Popconfirm>
      </Space>
    ),
  };

  return (
    <Table
      dataSource={data}
      columns={[...columns, actionColumn]}
      loading={isLoading}
      rowKey="id"
      pagination={{ pageSize: 20 }}
    />
  );
}
```

## 2. Form in Modal

```jsx
import { Modal, Form, Input, Select, message } from 'antd';
import { useMutation, useQueryClient } from '@tanstack/react-query';

export default function CreateModal({ open, onClose, createFn, queryKey }) {
  const [form] = Form.useForm();
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: createFn,
    onSuccess: () => {
      message.success('Created');
      queryClient.invalidateQueries({ queryKey });
      form.resetFields();
      onClose();
    },
  });

  return (
    <Modal
      title="Create"
      open={open}
      onOk={() => form.validateFields().then(mutation.mutate)}
      onCancel={onClose}
      confirmLoading={mutation.isPending}
    >
      <Form form={form} layout="vertical">
        <Form.Item name="name" label="Name" rules={[{ required: true }]}>
          <Input />
        </Form.Item>
      </Form>
    </Modal>
  );
}
```

## 3. Real-time Status Badge

```jsx
import { Badge, Tag } from 'antd';

const statusConfig = {
  running:  { color: 'green',  text: 'Running' },
  stopped:  { color: 'red',    text: 'Stopped' },
  pending:  { color: 'orange', text: 'Pending' },
  resolved: { color: 'blue',   text: 'Resolved' },
};

export function StatusTag({ status }) {
  const config = statusConfig[status] || { color: 'default', text: status };
  return <Tag color={config.color}>{config.text}</Tag>;
}
```

## 4. Chart Pattern (Recharts)

```jsx
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

export function DensityChart({ data }) {
  return (
    <ResponsiveContainer width="100%" height={300}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="hour" />
        <YAxis />
        <Tooltip />
        <Line type="monotone" dataKey="count" stroke="#1677ff" strokeWidth={2} />
      </LineChart>
    </ResponsiveContainer>
  );
}
```

## 5. Canvas Grid (Konva.js)

```jsx
import { Stage, Layer, Rect, Circle, Line, Text } from 'react-konva';

export function GridCanvas({ nodes, edges, width, height, cellSize = 20 }) {
  return (
    <Stage width={width} height={height}>
      <Layer>
        {/* Grid lines */}
        {Array.from({ length: Math.ceil(width / cellSize) }).map((_, i) => (
          <Line key={`v${i}`} points={[i * cellSize, 0, i * cellSize, height]}
                stroke="#f0f0f0" strokeWidth={1} />
        ))}

        {/* Edges */}
        {edges.map((e) => (
          <Line key={e.edge_id}
                points={[e.from_col * cellSize, e.from_row * cellSize,
                         e.to_col * cellSize, e.to_row * cellSize]}
                stroke="#91caff" strokeWidth={2} />
        ))}

        {/* Nodes */}
        {nodes.map((n) => (
          <Circle key={n.node_id}
                  x={n.col * cellSize} y={n.row * cellSize}
                  radius={6} fill="#1677ff"
                  onClick={() => console.log('clicked', n)} />
        ))}
      </Layer>
    </Stage>
  );
}
```

## 6. WebSocket Pattern

```jsx
import { useEffect, useRef } from 'react';
import useAuthStore from '../stores/authStore';

export function useWebSocket(path, onMessage) {
  const ws = useRef(null);
  const token = useAuthStore((s) => s.token);

  useEffect(() => {
    const url = `${import.meta.env.VITE_WS_URL}${path}?token=${token}`;
    ws.current = new WebSocket(url);
    
    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      onMessage(data);
    };
    
    ws.current.onerror = (err) => console.error('WS error:', err);
    
    return () => ws.current?.close();
  }, [path, token]);

  const send = (data) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(data));
    }
  };

  return { send };
}
```
