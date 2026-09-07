import { useEffect, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Card,
  CardActionArea,
  CardContent,
  Chip,
  CircularProgress,
  Divider,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import AddIcon from "@mui/icons-material/Add";
import SearchIcon from "@mui/icons-material/Search";
import { useNavigate } from "react-router-dom";
import { getLatestReports, toggleReportLike, type Report } from "../api";
import { useLocale } from "../i18n/LocaleContext";
import { LikeButton } from "../components/LikeButton";

export function AddReportStartPage() {
  const [productName, setProductName] = useState("");
  const [matches, setMatches] = useState<Report[]>([]);
  const [searched, setSearched] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { t, locale } = useLocale();
  const navigate = useNavigate();

  useEffect(() => {
    const trimmed = productName.trim();
    if (!trimmed) {
      setMatches([]);
      setSearched(false);
      return;
    }
    const timer = window.setTimeout(() => {
      setLoading(true);
      void getLatestReports(trimmed, locale, 10, 0)
        .then((page) => setMatches(page.reports))
        .catch(() => setError(t("addReport.searchError")))
        .finally(() => {
          setSearched(true);
          setLoading(false);
        });
    }, 300);
    return () => window.clearTimeout(timer);
  }, [productName, locale, t]);

  const toggleLike = async (reportId: number) => {
    try {
      const result = await toggleReportLike(reportId);
      setMatches((items) =>
        items.map((item) =>
          item.id === reportId
            ? { ...item, likedByMe: result.liked, likeCount: result.count }
            : item,
        ),
      );
    } catch {
      setError(t("reportDetail.likeError"));
    }
  };

  const addNewEntry = () => {
    const params = productName.trim() ? `?productName=${encodeURIComponent(productName.trim())}` : "";
    navigate(`/submit/new${params}`);
  };

  return (
    <Stack spacing={3} sx={{ maxWidth: 760 }}>
      <Box>
        <Typography variant="overline" color="secondary" sx={{ fontWeight: 700, letterSpacing: "0.14em" }}>
          {t("addReport.overline")}
        </Typography>
        <Typography variant="h2" sx={{ fontSize: { xs: "2.2rem", md: "3rem" } }}>
          {t("addReport.title")}
        </Typography>
        <Typography color="text.secondary">{t("addReport.subtitle")}</Typography>
      </Box>
      {error && <Alert severity="error">{error}</Alert>}
      <TextField
        autoFocus
        fullWidth
        label={t("newSubmission.productName")}
        placeholder={t("addReport.productNamePlaceholder")}
        value={productName}
        onChange={(event) => setProductName(event.target.value)}
        slotProps={{ input: { startAdornment: <SearchIcon sx={{ mr: 1, color: "text.secondary" }} /> } }}
      />
      {loading && (
        <Box sx={{ display: "grid", placeItems: "center", py: 4 }}>
          <CircularProgress size={28} />
        </Box>
      )}
      {!loading && searched && matches.length > 0 && (
        <Stack spacing={2}>
          <Typography color="text.secondary">{t("addReport.matchesFound")}</Typography>
          <Stack spacing={1.5}>
            {matches.map((report) => (
              <Card key={report.id} variant="outlined">
                <CardActionArea onClick={() => navigate(`/reports/${report.id}`)}>
                  <CardContent>
                    <Stack spacing={0.5}>
                      <Typography variant="h6">{report.product}</Typography>
                      <Typography color="text.secondary">{report.brand}</Typography>
                      <Stack direction="row" spacing={1} sx={{ flexWrap: "wrap" }}>
                        <Chip label={report.category} size="small" />
                        <Chip label={report.seller} size="small" variant="outlined" />
                      </Stack>
                    </Stack>
                  </CardContent>
                </CardActionArea>
                <Box sx={{ px: 2, pb: 1.5 }}>
                  <LikeButton
                    liked={report.likedByMe}
                    count={report.likeCount}
                    onToggle={() => void toggleLike(report.id)}
                    label={t("reportDetail.like")}
                  />
                </Box>
              </Card>
            ))}
          </Stack>
        </Stack>
      )}
      {!loading && searched && matches.length === 0 && (
        <Alert severity="info">{t("addReport.noMatches")}</Alert>
      )}
      {searched && !loading && (
        <>
          <Divider />
          <Box>
            <Typography color="text.secondary" sx={{ mb: 1.5 }}>
              {matches.length > 0 ? t("addReport.notWhatYouMeant") : t("addReport.startFresh")}
            </Typography>
            <Button variant="contained" size="large" startIcon={<AddIcon />} onClick={addNewEntry}>
              {t("addReport.addNewEntry")}
            </Button>
          </Box>
        </>
      )}
    </Stack>
  );
}
