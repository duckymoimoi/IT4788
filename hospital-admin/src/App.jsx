import { Routes, Route } from 'react-router-dom';
import AdminLayout from './components/Layout/AdminLayout';
import AuthGuard from './components/Layout/AuthGuard';

// A — Nền tảng
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';

// B — Map & Engine
import MapEditor from './pages/MapEditor';
import MapManager from './pages/MapManager';
import MapBuilder from './pages/MapBuilder';
import EnginePanel from './pages/EnginePanel';

// C — Flow & Simulation
import FlowMonitor from './pages/FlowMonitor';
import SimControl from './pages/SimControl';

// D — Medical, Device & Settings
import MedicalDash from './pages/MedicalDash';
import DeviceMonitor from './pages/DeviceMonitor';
import SystemSettings from './pages/SystemSettings';

// E — SOS & Chat
import SOSDashboard from './pages/SOSDashboard';
import ChatSupport from './pages/ChatSupport';

function App() {
  return (
    <Routes>
      {/* Public */}
      <Route path="/login" element={<Login />} />

      {/* Protected — requires login */}
      <Route
        path="/"
        element={
          <AuthGuard>
            <AdminLayout />
          </AuthGuard>
        }
      >
        <Route index element={<Dashboard />} />
        <Route path="map-editor" element={<MapEditor />} />
        <Route path="map-manager" element={<MapManager />} />
        <Route path="map-builder" element={<MapBuilder />} />
        <Route path="flow" element={<FlowMonitor />} />
        <Route path="sim" element={<SimControl />} />
        <Route path="medical" element={<MedicalDash />} />
        <Route path="device" element={<DeviceMonitor />} />
        <Route path="sos" element={<SOSDashboard />} />
        <Route path="chat" element={<ChatSupport />} />
        <Route path="engine" element={<EnginePanel />} />
        <Route path="settings" element={<SystemSettings />} />
      </Route>
    </Routes>
  );
}

export default App;
