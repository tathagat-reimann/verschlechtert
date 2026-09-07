import { useState } from 'react'
import { Avatar, Box, Button, Container, Divider, IconButton, Menu, MenuItem, Stack, Toolbar, ToggleButton, ToggleButtonGroup, Typography, AppBar } from '@mui/material'
import MenuIcon from '@mui/icons-material/Menu'
import { Link as RouterLink, Route, Routes } from 'react-router-dom'
import { useAuth } from './auth/AuthContext'
import { useLocale } from './i18n/LocaleContext'
import { ProductHomePage } from './pages/ProductHomePage'
import { NewSubmissionPage } from './pages/NewSubmissionPage'
import { MySubmissionsPage } from './pages/MySubmissionsPage'
import { ReportDetailPage } from './pages/ReportDetailPage'

function App() {
  const { user, signOutUser } = useAuth()
  const { locale, setLocale, t } = useLocale()
  const [menuAnchor, setMenuAnchor] = useState<HTMLElement | null>(null)
  const closeMenu = () => setMenuAnchor(null)

  return (
    <Box sx={{ minHeight: "100svh", bgcolor: "background.default" }}>
      <AppBar position="static" color="transparent" elevation={0} sx={{ borderBottom: "1px solid", borderColor: "divider" }}>
        <Container maxWidth="lg">
          <Toolbar disableGutters sx={{ justifyContent: "space-between", gap: 2 }}>
            <Stack direction="row" spacing={1} sx={{ alignItems: "center" }}>
              <IconButton color="inherit" aria-label="Menu" sx={{ display: { xs: "inline-flex", sm: "none" } }} onClick={(event) => setMenuAnchor(event.currentTarget)}>
                <MenuIcon />
              </IconButton>
              <Typography component={RouterLink} to="/" variant="h6" color="primary" sx={{ fontWeight: 800, textDecoration: "none" }}>Verschlechtert</Typography>
            </Stack>
            <Stack direction="row" spacing={1.5} sx={{ alignItems: "center" }}>
              <ToggleButtonGroup value={locale} exclusive size="small" onChange={(_, value) => value && setLocale(value)} aria-label="Language" sx={{ display: { xs: "none", sm: "inline-flex" } }}>
                <ToggleButton value="de" aria-label="Deutsch">DE</ToggleButton>
                <ToggleButton value="en" aria-label="English">EN</ToggleButton>
              </ToggleButtonGroup>
              <Button component={RouterLink} to="/submissions" color="inherit" sx={{ display: { xs: "none", sm: "inline-flex" } }}>{t("header.mySubmissions")}</Button>
              <Avatar src={user?.photoURL ?? undefined} alt={user?.displayName ?? "User"} sx={{ width: 32, height: 32 }} />
              <Typography sx={{ display: { xs: "none", sm: "block" } }}>{user?.displayName ?? user?.email}</Typography>
              <Button onClick={() => void signOutUser()} color="inherit" sx={{ display: { xs: "none", sm: "inline-flex" } }}>{t("header.signOut")}</Button>
            </Stack>
          </Toolbar>
        </Container>
      </AppBar>
      <Menu anchorEl={menuAnchor} open={Boolean(menuAnchor)} onClose={closeMenu}>
        <MenuItem component={RouterLink} to="/submissions" onClick={closeMenu}>{t("header.mySubmissions")}</MenuItem>
        <Divider />
        <MenuItem selected={locale === "de"} onClick={() => { setLocale("de"); closeMenu(); }}>Deutsch (DE)</MenuItem>
        <MenuItem selected={locale === "en"} onClick={() => { setLocale("en"); closeMenu(); }}>English (EN)</MenuItem>
        <Divider />
        <MenuItem onClick={() => { closeMenu(); void signOutUser(); }}>{t("header.signOut")}</MenuItem>
      </Menu>
      <Container maxWidth="lg" sx={{ py: { xs: 6, md: 10 } }}>
        <Routes>
          <Route path="/" element={<ProductHomePage />} />
          <Route path="/submit" element={<NewSubmissionPage />} />
          <Route path="/submissions" element={<MySubmissionsPage />} />
          <Route path="/reports/:id" element={<ReportDetailPage />} />
        </Routes>
      </Container>
    </Box>
  )
}

export default App
