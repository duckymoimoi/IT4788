import { useState, useEffect, useRef, useCallback } from 'react';
import {
  Typography, Row, Col, Card, Button, Space, InputNumber, Tag, Descriptions,
  message, Spin, Slider, Divider, Statistic, Form, Select, Alert, Tooltip,
} from 'antd';
import {
  SettingOutlined, PlayCircleOutlined, PauseCircleOutlined,
  StepForwardOutlined, StepBackwardOutlined, ThunderboltOutlined,
  DeleteOutlined, CheckCircleOutlined, CloseCircleOutlined,
  ReloadOutlined, RocketOutlined, AimOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  fetchEngineHealth, solve, setParams, fetchConvergence,
  clearCache, loadMapf, fetchMapfPositions, fetchMapfInfo,
} from '../api/engine';
import { fetchFloors, fetchNodes, fetchMeta } from '../api/map';
import GridCanvas from '../components/GridCanvas/GridCanvas';

const { Title, Text } = Typography;

// ─── Health Status Badge ─────────────────────────────────────
function StatusBadge({ ok, label }) {
  return (
    <Space size={4}>
      {ok ? (
        <CheckCircleOutlined style={{ color: '#52c41a' }} />
      ) : (
        <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
      )}
      <Text style={{ fontSize: 13 }}>{label}</Text>
    </Space>
  );
}

// ─── B7: Pathfinding Test Section ────────────────────────────
function PathfindingSection({ nodes, meta }) {
  const [startLoc, setStartLoc] = useState(null);
  const [destLoc, setDestLoc] = useState(null);
  const [pathResult, setPathResult] = useState(null);

  const solveMutation = useMutation({
    mutationFn: solve,
    onSuccess: (data) => {
      setPathResult(data);
      message.success(`Tìm đường thành công! Distance: ${data?.total_distance ?? data?.distance ?? '?'}`);
    },
    onError: (err) => {
      message.error(`Lỗi: ${err.response?.data?.message || err.message}`);
      setPathResult(null);
    },
  });

  const handleSolve = () => {
    if (!startLoc || !destLoc) {
      message.warning('Chọn điểm bắt đầu và điểm đến');
      return;
    }
    if (startLoc === destLoc) {
      message.warning('Điểm bắt đầu và đích không được trùng nhau');
      return;
    }
    solveMutation.mutate({
      start_location: startLoc,
      dest_location: destLoc,
      mode_id: 'walking',
    });
  };

  // Build options from nodes
  const nodeOptions = (nodes || []).map((n) => ({
    value: n.grid_location,
    label: `${n.poi_code} — ${n.poi_name} (${n.grid_location})`,
  }));

  // Extract path cells for highlighting
  const pathCells = pathResult?.path || pathResult?.steps?.map((s) => s.grid_location) || [];

  return (
    <Card
      title={
        <Space>
          <RocketOutlined style={{ color: '#1677ff' }} />
          <span>Pathfinding Test (Dijkstra)</span>
        </Space>
      }
    >
      <Row gutter={16} align="middle">
        <Col flex="1">
          <Space direction="vertical" style={{ width: '100%' }}>
            <Text strong style={{ fontSize: 12 }}>Điểm bắt đầu:</Text>
            <Select
              showSearch
              value={startLoc}
              onChange={setStartLoc}
              placeholder="Chọn POI bắt đầu..."
              options={nodeOptions}
              style={{ width: '100%' }}
              filterOption={(input, option) =>
                option.label.toLowerCase().includes(input.toLowerCase())
              }
            />
          </Space>
        </Col>
        <Col>
          <Text style={{ fontSize: 20, color: '#bfbfbf', marginTop: 20, display: 'block' }}>→</Text>
        </Col>
        <Col flex="1">
          <Space direction="vertical" style={{ width: '100%' }}>
            <Text strong style={{ fontSize: 12 }}>Điểm đến:</Text>
            <Select
              showSearch
              value={destLoc}
              onChange={setDestLoc}
              placeholder="Chọn POI đích..."
              options={nodeOptions}
              style={{ width: '100%' }}
              filterOption={(input, option) =>
                option.label.toLowerCase().includes(input.toLowerCase())
              }
            />
          </Space>
        </Col>
        <Col>
          <Button
            type="primary"
            icon={<ThunderboltOutlined />}
            onClick={handleSolve}
            loading={solveMutation.isPending}
            size="large"
            style={{ marginTop: 18 }}
          >
            Solve
          </Button>
        </Col>
      </Row>

      {/* Result */}
      {pathResult && (
        <div style={{ marginTop: 16 }}>
          <Alert
            type="success"
            showIcon
            message={
              <Space split={<Divider type="vertical" />}>
                <span>
                  Distance: <Text strong>{pathResult.total_distance ?? pathResult.distance ?? '?'}</Text>
                </span>
                <span>
                  Steps: <Text strong>{pathCells.length}</Text>
                </span>
                {pathResult.estimated_time && (
                  <span>
                    ETA: <Text strong>~{pathResult.estimated_time}s</Text>
                  </span>
                )}
              </Space>
            }
            style={{ marginBottom: 12 }}
          />

          {/* Path visualization on grid */}
          <GridCanvas
            rows={meta?.rows || 33}
            cols={meta?.cols || 57}
            gridData={meta?.grid_data}
            nodes={nodes || []}
            pathCells={pathCells}
            width={Math.min(900, window.innerWidth - 400)}
            height={420}
          />

          {/* Path cell list */}
          <div style={{ marginTop: 8, maxHeight: 60, overflow: 'auto' }}>
            <Text style={{ fontSize: 11, color: '#999' }}>
              Path: {pathCells.slice(0, 30).join(' → ')}
              {pathCells.length > 30 && ` ... (+${pathCells.length - 30} more)`}
            </Text>
          </div>
        </div>
      )}
    </Card>
  );
}

// ─── B8: MAPF Viewer Section ─────────────────────────────────
function MAPFViewerSection({ nodes, meta }) {
  const [timestep, setTimestep] = useState(0);
  const [playing, setPlaying] = useState(false);
  const [playSpeed, setPlaySpeed] = useState(500);
  const intervalRef = useRef(null);
  const queryClient = useQueryClient();

  // Fetch MAPF info
  const { data: mapfInfo, isLoading: loadingInfo, refetch: refetchInfo } = useQuery({
    queryKey: ['mapf-info'],
    queryFn: fetchMapfInfo,
  });

  const makespan = mapfInfo?.makespan || mapfInfo?.max_timestep || 0;
  const agentCount = mapfInfo?.agent_count || mapfInfo?.team_size || 0;

  // Fetch positions for current timestep
  const { data: positions, isLoading: loadingPos } = useQuery({
    queryKey: ['mapf-positions', timestep],
    queryFn: () => fetchMapfPositions(timestep),
    enabled: makespan > 0,
  });

  // Load MAPF mutation
  const loadMutation = useMutation({
    mutationFn: loadMapf,
    onSuccess: () => {
      message.success('MAPF data loaded!');
      refetchInfo();
      setTimestep(0);
    },
    onError: (err) => {
      message.error(`Load MAPF thất bại: ${err.response?.data?.message || err.message}`);
    },
  });

  // Auto-play
  useEffect(() => {
    if (playing && makespan > 0) {
      intervalRef.current = setInterval(() => {
        setTimestep((prev) => {
          if (prev >= makespan) {
            setPlaying(false);
            return prev;
          }
          return prev + 1;
        });
      }, playSpeed);
    }
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [playing, makespan, playSpeed]);

  // Parse agent positions
  const agentPositions = Array.isArray(positions)
    ? positions
    : positions?.agents || positions?.positions || [];

  return (
    <Card
      title={
        <Space>
          <PlayCircleOutlined style={{ color: '#722ed1' }} />
          <span>MAPF Viewer</span>
          {agentCount > 0 && <Tag color="purple">{agentCount} agents</Tag>}
          {makespan > 0 && <Tag color="cyan">Makespan: {makespan}</Tag>}
        </Space>
      }
      extra={
        <Button
          onClick={() => loadMutation.mutate('output.json')}
          loading={loadMutation.isPending}
          icon={<ReloadOutlined />}
          size="small"
        >
          Load MAPF
        </Button>
      }
      style={{ marginTop: 16 }}
    >
      {makespan === 0 ? (
        <Alert
          type="info"
          showIcon
          message="Chưa có dữ liệu MAPF"
          description="Nhấn 'Load MAPF' để tải dữ liệu pre-computed paths, hoặc upload file output.json mới."
        />
      ) : (
        <>
          {/* Timestep controls */}
          <Row gutter={16} align="middle" style={{ marginBottom: 16 }}>
            <Col flex="auto">
              <div style={{ padding: '0 8px' }}>
                <Text strong style={{ fontSize: 12, display: 'block', marginBottom: 4 }}>
                  Timestep: {timestep} / {makespan}
                </Text>
                <Slider
                  min={0}
                  max={makespan}
                  value={timestep}
                  onChange={setTimestep}
                  tooltip={{ formatter: (v) => `T = ${v}` }}
                />
              </div>
            </Col>
            <Col>
              <Space>
                <Tooltip title="Previous">
                  <Button
                    icon={<StepBackwardOutlined />}
                    onClick={() => setTimestep(Math.max(0, timestep - 1))}
                    disabled={timestep <= 0}
                  />
                </Tooltip>
                <Button
                  type={playing ? 'default' : 'primary'}
                  icon={playing ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                  onClick={() => setPlaying(!playing)}
                  danger={playing}
                >
                  {playing ? 'Pause' : 'Play'}
                </Button>
                <Tooltip title="Next">
                  <Button
                    icon={<StepForwardOutlined />}
                    onClick={() => setTimestep(Math.min(makespan, timestep + 1))}
                    disabled={timestep >= makespan}
                  />
                </Tooltip>
                <Select
                  value={playSpeed}
                  onChange={setPlaySpeed}
                  style={{ width: 100 }}
                  size="small"
                  options={[
                    { value: 200, label: '0.2s (5x)' },
                    { value: 500, label: '0.5s (2x)' },
                    { value: 1000, label: '1s (1x)' },
                    { value: 2000, label: '2s (0.5x)' },
                  ]}
                />
              </Space>
            </Col>
          </Row>

          {/* Canvas with agents */}
          <GridCanvas
            rows={meta?.rows || 33}
            cols={meta?.cols || 57}
            gridData={meta?.grid_data}
            nodes={nodes || []}
            agentPositions={agentPositions}
            width={Math.min(900, window.innerWidth - 400)}
            height={420}
          />
        </>
      )}
    </Card>
  );
}

// ─── B9: Engine Controls Section ─────────────────────────────
function EngineControlsSection() {
  const [form] = Form.useForm();
  const queryClient = useQueryClient();

  // Health
  const { data: health, isLoading: loadingHealth } = useQuery({
    queryKey: ['engine-health'],
    queryFn: fetchEngineHealth,
    refetchInterval: 15000,
  });

  // Convergence
  const { data: convergence } = useQuery({
    queryKey: ['convergence'],
    queryFn: fetchConvergence,
    refetchInterval: 15000,
  });

  // Set params mutation
  const paramsMutation = useMutation({
    mutationFn: setParams,
    onSuccess: () => {
      message.success('Cập nhật tham số thành công!');
      queryClient.invalidateQueries({ queryKey: ['engine-health'] });
    },
    onError: (err) => {
      message.error(`Lỗi: ${err.response?.data?.message || err.message}`);
    },
  });

  // Clear cache mutation
  const cacheMutation = useMutation({
    mutationFn: clearCache,
    onSuccess: () => {
      message.success('Cache đã được xóa!');
      queryClient.invalidateQueries({ queryKey: ['engine-health'] });
    },
    onError: (err) => {
      message.error(`Lỗi: ${err.response?.data?.message || err.message}`);
    },
  });

  const handleSaveParams = async () => {
    const values = await form.validateFields();
    paramsMutation.mutate(values);
  };

  return (
    <Row gutter={16} style={{ marginTop: 16 }}>
      {/* Engine Status */}
      <Col xs={24} md={8}>
        <Card
          title={
            <Space>
              <SettingOutlined />
              <span>Engine Status</span>
            </Space>
          }
          loading={loadingHealth}
        >
          {health ? (
            <Space direction="vertical" size={8} style={{ width: '100%' }}>
              <StatusBadge ok={health.status === 'ok'} label={`Status: ${health.status || 'unknown'}`} />
              <StatusBadge ok={health.db_connected} label="Database" />
              <StatusBadge ok={health.grid_loaded} label="Grid Loaded" />
              <StatusBadge ok={health.mapf_loaded} label="MAPF Loaded" />
              <Divider style={{ margin: '8px 0' }} />
              <Statistic
                title="Agent Count"
                value={health.agent_count || 0}
                valueStyle={{ fontSize: 20 }}
              />
              {convergence != null && (
                <Statistic
                  title="Convergence"
                  value={typeof convergence === 'number' ? convergence : convergence?.value ?? 0}
                  suffix="%"
                  precision={1}
                  valueStyle={{ fontSize: 20, color: '#52c41a' }}
                />
              )}
            </Space>
          ) : (
            <Alert type="warning" message="Không thể kết nối Engine" />
          )}
        </Card>
      </Col>

      {/* Engine Parameters */}
      <Col xs={24} md={8}>
        <Card
          title="⚙️ Tham số Engine"
          extra={
            <Button
              type="primary"
              size="small"
              onClick={handleSaveParams}
              loading={paramsMutation.isPending}
            >
              Lưu
            </Button>
          }
        >
          <Form form={form} layout="vertical" size="small">
            <Form.Item label="Alpha (α — trọng số khoảng cách)" name="alpha" initialValue={1.0}>
              <InputNumber min={0} max={10} step={0.1} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item label="Beta (β — trọng số mật độ)" name="beta" initialValue={0.5}>
              <InputNumber min={0} max={10} step={0.1} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item label="Max iterations" name="max_iterations" initialValue={1000}>
              <InputNumber min={100} max={100000} step={100} style={{ width: '100%' }} />
            </Form.Item>
          </Form>
        </Card>
      </Col>

      {/* Cache Management */}
      <Col xs={24} md={8}>
        <Card title="🗑️ Cache & Maintenance">
          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            <Alert
              type="info"
              showIcon
              message="Dijkstra cache"
              description="Xóa cache để buộc tính lại tất cả đường đi. Thường dùng sau khi thay đổi tham số hoặc upload map mới."
              style={{ fontSize: 12 }}
            />
            <Button
              danger
              icon={<DeleteOutlined />}
              onClick={() => cacheMutation.mutate()}
              loading={cacheMutation.isPending}
              block
              size="large"
            >
              Clear All Cache
            </Button>
            <Button
              icon={<ReloadOutlined />}
              onClick={() => {
                queryClient.invalidateQueries({ queryKey: ['engine-health'] });
                queryClient.invalidateQueries({ queryKey: ['convergence'] });
                message.info('Refreshing...');
              }}
              block
            >
              Refresh Status
            </Button>
          </Space>
        </Card>
      </Col>
    </Row>
  );
}

// ─── Engine Panel Page ────────────────────────────────────────
export default function EnginePanel() {
  // Load floor + nodes for pathfinding visualization
  const { data: floors } = useQuery({
    queryKey: ['floors'],
    queryFn: fetchFloors,
  });

  const activeFloor = floors?.[0]?.map_id ?? null;

  const { data: meta } = useQuery({
    queryKey: ['meta', activeFloor],
    queryFn: () => fetchMeta(activeFloor),
    enabled: !!activeFloor,
  });

  const { data: nodes } = useQuery({
    queryKey: ['nodes', activeFloor],
    queryFn: () => fetchNodes(activeFloor),
    enabled: !!activeFloor,
  });

  return (
    <>
      <Title level={4}>
        <SettingOutlined style={{ marginRight: 8 }} />
        Engine Panel
      </Title>

      {/* B7: Pathfinding Test */}
      <PathfindingSection nodes={nodes} meta={meta} />

      {/* B8: MAPF Viewer */}
      <MAPFViewerSection nodes={nodes} meta={meta} />

      {/* B9: Controls & Status */}
      <EngineControlsSection />
    </>
  );
}
