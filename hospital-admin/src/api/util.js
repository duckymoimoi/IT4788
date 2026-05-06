import api from './client';

export const fetchFAQ = () => api.get('/util/faq').then((r) => r.data.data);
export const fetchAbout = () => api.get('/util/about').then((r) => r.data.data);
export const fetchContact = () => api.get('/util/contact').then((r) => r.data.data);
export const fetchFeedbackSummary = () => api.get('/util/feedback_summary').then((r) => r.data.data);
export const fetchNotifications = () => api.get('/notification/get_list').then((r) => r.data.data);
export const deleteNotification = (id) => api.delete('/notification/delete', { data: { id } }).then((r) => r.data);
export const checkVersion = () => api.get('/sys/check_version').then((r) => r.data.data);
