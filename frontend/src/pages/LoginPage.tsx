import { useState } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";

export function LoginPage() {
  const { user, signInWithGoogle } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();
  const location = useLocation();

  const redirectTo = (location.state as { from?: Location })?.from?.pathname ?? "/";

  if (user) {
    return <Navigate to={redirectTo} replace />;
  }

  const handleSignIn = async () => {
    setError(null);
    try {
      await signInWithGoogle();
      navigate(redirectTo, { replace: true });
    } catch {
      setError("Sign-in failed. Please try again.");
    }
  };

  return (
    <section style={{ maxWidth: 360, margin: "4rem auto", textAlign: "center" }}>
      <h1>Sign in</h1>
      <p>Sign in with your Google account to continue.</p>
      <button type="button" onClick={handleSignIn}>
        Sign in with Google
      </button>
      {error && <p role="alert">{error}</p>}
    </section>
  );
}
