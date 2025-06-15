import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import Layout from "./components/ui/layout";
import MainPage from "./components/MainPage";
import { Toaster } from 'sonner';
import LoginPage from './pages/login';
import SignUpPage from './pages/signup';
import OnboardingPage from './pages/onboarding';
import IntegrationsPage from './pages/integrations/IntegrationsPage';
import IntegrationDetailsPage from './pages/integrations/IntegrationDetailsPage';
import IntegrationCallbackPage from './pages/integrations/IntegrationCallbackPage';
import { RedirectToSignIn, useAuth } from '@clerk/clerk-react';
import { useOnboardingGuard } from './hooks/useOnboardingGuard';

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { isSignedIn, isLoaded } = useAuth();
  const { isLoading, isComplete } = useOnboardingGuard();
  const location = useLocation();

  // Wait for Clerk to load
  if (!isLoaded) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  // Redirect to sign in if not authenticated
  if (!isSignedIn) {
    return <RedirectToSignIn />;
  }

  // Show loading while checking onboarding status
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Checking onboarding status...</p>
        </div>
      </div>
    );
  }

  // Redirect to onboarding if not complete (covers both org creation and metadata)
  if (!isComplete && location.pathname !== '/onboarding') {
    return <Navigate to="/onboarding" replace />;
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
        <Route path="/onboarding" element={
          <ProtectedRoute>
              <OnboardingPage />
          </ProtectedRoute>
        } />
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
                <IntegrationsPage />
              </Layout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/integrations/:connectorType"
          element={
            <ProtectedRoute>
              <Layout>
                <IntegrationDetailsPage />
              </Layout>
            </ProtectedRoute>
          }
        />
        <Route
          path="/integrations/:connectorType/authorize"
          element={<IntegrationCallbackPage />}
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
