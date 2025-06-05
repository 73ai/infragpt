import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Layout from "./components/ui/layout";
import MainPage from "./components/MainPage";
import { Toaster } from 'sonner';
import LoginPage from './pages/login';
import SignUpPage from './pages/signup';
import { RedirectToSignIn, useAuth } from '@clerk/clerk-react';

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { isSignedIn } = useAuth();
  if (!isSignedIn) {
    const currentUrl = window.location.href;
    if (!currentUrl.includes('signup')) {
      return <RedirectToSignIn />;
    }
  }
  return <>{children}</>;
}

const App = () => {

  return (
    <>
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/signup" element={<SignUpPage />} />
        <Route
          path="/dashboard"
          element={
            <ProtectedRoute>
              <Layout>
                <MainPage />
              </Layout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/integrations"
          element={
            <ProtectedRoute>
              <Layout>
                <MainPage />
              </Layout>
            </ProtectedRoute>
          }
        />
        {/* Redirect root to dashboard */}
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        {/* Catch all route - 404 */}
        <Route
          path="*"
          element={
            <div className="flex items-center justify-center h-screen">
              <div className="text-center">
                <h1 className="text-4xl font-bold mb-4">404</h1>
                <p className="text-muted-foreground">Page not found</p>
              </div>
            </div>
          }
        />
      </Routes>
      <Toaster position="top-right" richColors />
    </BrowserRouter>
    </>
  );
};

export default App;
