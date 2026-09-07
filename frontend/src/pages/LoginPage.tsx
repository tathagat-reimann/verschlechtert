import { useState } from "react";
import { Alert, Box, Button, Paper, Stack, Typography } from "@mui/material";
import GoogleIcon from "@mui/icons-material/Google";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { useLocale } from "../i18n/LocaleContext";

export function LoginPage() {
  const { user, signInWithGoogle } = useAuth();
  const { t } = useLocale();
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
      setError(t("login.signInFailed"));
    }
  };

  return (
    <Box sx={{ minHeight: "100svh", display: "grid", placeItems: "center", p: 3 }}>
      <Paper elevation={0} sx={{ width: "min(100%, 420px)", p: { xs: 3, sm: 5 }, border: "1px solid", borderColor: "divider" }}>
        <Stack spacing={3}>
          <Box>
            <Typography color="primary" sx={{ fontWeight: 700, letterSpacing: "0.12em", textTransform: "uppercase" }} variant="overline">
              {t("login.tagline")}
            </Typography>
            <Typography variant="h1" sx={{ fontSize: { xs: "2.5rem", sm: "3.5rem" }, mt: 1 }}>
              {t("login.title")}
            </Typography>
            <Typography color="text.secondary" sx={{ mt: 2 }}>
              {t("login.subtitle")}
            </Typography>
          </Box>
          {error && <Alert severity="error" role="alert">{error}</Alert>}
          <Button onClick={handleSignIn} startIcon={<GoogleIcon />} variant="contained" size="large" fullWidth>
            {t("login.continueWithGoogle")}
          </Button>
        </Stack>
      </Paper>
    </Box>
  );
}
