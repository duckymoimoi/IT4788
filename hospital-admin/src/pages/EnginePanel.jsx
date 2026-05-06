import { Typography } from 'antd';

const { Title, Text } = Typography;

export default function EnginePanel() {
  return (
    <>
      <Title level={4}>Engine Panel</Title>
      <Text type="secondary">👤 B — Pathfinding, MAPF viewer, params</Text>
    </>
  );
}
