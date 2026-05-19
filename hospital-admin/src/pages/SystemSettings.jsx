import React from 'react';
import { Typography, Card, Tabs, Table, Button, Collapse, Descriptions, Statistic, Row, Col, Space, message, List } from 'antd';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  fetchNotifications,
  deleteNotification,
  fetchFeedbackSummary,
  fetchFAQ,
  fetchAbout
} from '../api/util';

const { Title, Text, Paragraph } = Typography;

export default function SystemSettings() {
  const queryClient = useQueryClient();

  // Queries
  const { data: notifications, isLoading: notifLoading } = useQuery({
    queryKey: ['notifications'],
    queryFn: fetchNotifications,
  });

  const { data: feedback, isLoading: feedbackLoading } = useQuery({
    queryKey: ['feedbackSummary'],
    queryFn: fetchFeedbackSummary,
  });

  const { data: faqList, isLoading: faqLoading } = useQuery({
    queryKey: ['faq'],
    queryFn: fetchFAQ,
  });

  const { data: aboutData, isLoading: aboutLoading } = useQuery({
    queryKey: ['about'],
    queryFn: fetchAbout,
  });

  // Mutations
  const delNotifMutation = useMutation({
    mutationFn: deleteNotification,
    onSuccess: () => {
      message.success('Notification deleted');
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
    onError: () => message.error('Failed to delete notification')
  });

  // Tabs Content
  const renderNotifications = () => {
    const columns = [
      { title: 'Type', dataIndex: 'type', key: 'type' },
      { title: 'Message', dataIndex: 'message', key: 'message' },
      { title: 'Date', dataIndex: 'created_at', key: 'created_at' },
      { title: 'Action', key: 'action', render: (_, record) => (
          <Button 
            type="link" 
            danger 
            onClick={() => delNotifMutation.mutate(record.id)}
            loading={delNotifMutation.isPending}
          >
            Delete
          </Button>
        ) 
      }
    ];

    return (
      <Table 
        dataSource={notifications || []} 
        columns={columns} 
        loading={notifLoading} 
        rowKey="id" 
        pagination={{ pageSize: 10 }}
      />
    );
  };

  const renderFeedback = () => {
    if (feedbackLoading) return <Text type="secondary">Loading...</Text>;
    if (!feedback) return <Text>No feedback available</Text>;

    return (
      <Row gutter={[16, 16]}>
        <Col span={8}>
          <Card>
            <Statistic title="Total Feedback" value={feedback.total || 0} />
          </Card>
        </Col>
        <Col span={8}>
          <Card>
            <Statistic title="Average Rating" value={feedback.average_rating || 0} suffix="/ 5" />
          </Card>
        </Col>
        <Col span={24}>
          <Title level={5}>Recent Comments</Title>
          <List
            bordered
            dataSource={feedback.recent_comments || []}
            renderItem={item => (
              <List.Item>
                <Text strong>{item.user}:</Text> {item.comment} <Text type="secondary">({item.rating}⭐)</Text>
              </List.Item>
            )}
          />
        </Col>
      </Row>
    );
  };

  const renderFAQ = () => {
    if (faqLoading) return <Text type="secondary">Loading...</Text>;
    
    // Ant Design Collapse items structure
    const items = (faqList || []).map((faq, index) => ({
      key: String(index),
      label: faq.question,
      children: <p>{faq.answer}</p>
    }));

    return items.length > 0 ? <Collapse items={items} /> : <Text>No FAQ available</Text>;
  };

  const renderAbout = () => {
    if (aboutLoading) return <Text type="secondary">Loading...</Text>;
    if (!aboutData) return <Text>No about information available</Text>;

    return (
      <Descriptions bordered column={1}>
        <Descriptions.Item label="App Name">{aboutData.app_name}</Descriptions.Item>
        <Descriptions.Item label="Version">{aboutData.version}</Descriptions.Item>
        <Descriptions.Item label="Developer">{aboutData.developer}</Descriptions.Item>
        <Descriptions.Item label="Contact Email">{aboutData.contact_email}</Descriptions.Item>
        <Descriptions.Item label="Description">{aboutData.description}</Descriptions.Item>
      </Descriptions>
    );
  };

  return (
    <>
      <Title level={4} style={{ marginBottom: 16 }}>System Settings</Title>
      
      <Card>
        <Tabs defaultActiveKey="1" items={[
          {
            key: '1',
            label: 'Notifications',
            children: renderNotifications()
          },
          {
            key: '2',
            label: 'Feedback',
            children: renderFeedback()
          },
          {
            key: '3',
            label: 'FAQ',
            children: renderFAQ()
          },
          {
            key: '4',
            label: 'About',
            children: renderAbout()
          }
        ]} />
      </Card>
    </>
  );
}
