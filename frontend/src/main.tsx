import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { CssBaseline, ThemeProvider, createTheme } from '@mui/material'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import './index.css'
import App from './App.tsx'
import { AuthProvider } from './auth/AuthContext.tsx'
import { LocaleProvider } from './i18n/LocaleContext.tsx'
import { ProtectedRoute } from './auth/ProtectedRoute.tsx'
import { LoginPage } from './pages/LoginPage.tsx'

const theme = createTheme({
  palette: {
    primary: { main: '#0f766e' },
    secondary: { main: '#d97706' },
    background: { default: '#f6f7f5', paper: '#ffffff' },
  },
  typography: {
    fontFamily: '"DM Sans", "Helvetica Neue", Arial, sans-serif',
    h1: { fontFamily: 'Georgia, serif', fontWeight: 600 },
    h2: { fontFamily: 'Georgia, serif', fontWeight: 600 },
  },
  shape: { borderRadius: 10 },
});

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <BrowserRouter>
        <AuthProvider>
          <LocaleProvider>
            <Routes>
              <Route path="/login" element={<LoginPage />} />
              <Route
                path="/*"
                element={
                  <ProtectedRoute>
                    <App />
                  </ProtectedRoute>
                }
              />
            </Routes>
          </LocaleProvider>
        </AuthProvider>
      </BrowserRouter>
    </ThemeProvider>
  </StrictMode>,
)
