import { Navigate } from 'react-router-dom';
import useAuthStore from '../../stores/authStore';

export default function AuthGuard({ children }) {
  const token = useAuthStore((s) => s.token);

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return children;
}
