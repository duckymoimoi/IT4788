import { Typography } from 'antd';

const { Title, Text } = Typography;

export default function MedicalDash() {
  return (
    <>
      <Title level={4}>Medical Dashboard</Title>
      <Text type="secondary">👤 D — Hàng đợi, chỉ định, kết quả xét nghiệm</Text>
    </>
  );
}
