import { useEffect, useRef, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Card,
  CardActionArea,
  CardContent,
  Chip,
  CircularProgress,
  Grid,
  InputAdornment,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import SearchIcon from "@mui/icons-material/Search";
import AddIcon from "@mui/icons-material/Add";
import { Link as RouterLink, useNavigate } from "react-router-dom";
import { getLatestReports, toggleReportLike, type Report } from "../api";
import { useLocale } from "../i18n/LocaleContext";
import { LikeButton } from "../components/LikeButton";

const PAGE_SIZE = 24;

export function ProductHomePage() {
  const [reports, setReports] = useState<Report[]>([]);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { t, locale } = useLocale();
  const navigate = useNavigate();
  const sentinelRef = useRef<HTMLDivElement | null>(null);
  // Track offset/in-flight state in refs (not state) so fast successive intersection
  // callbacks can't race ahead of React's state updates and fetch the same page twice.
  const offsetRef = useRef(0);
  const loadingMoreRef = useRef(false);

  // Reset and load the first page whenever the search query or locale changes.
  useEffect(() => {
    const timer = window.setTimeout(() => {
      setLoading(true);
      offsetRef.current = 0;
      void getLatestReports(query, locale, PAGE_SIZE, 0)
        .then((page) => {
          setReports(page.reports);
          setHasMore(page.hasMore);
          offsetRef.current = page.reports.length;
        })
        .catch(() => setError(t("home.loadError")))
        .finally(() => setLoading(false));
    }, 250);
    return () => window.clearTimeout(timer);
  }, [query, locale, t]);

  // Load the next page when the sentinel element at the bottom of the list becomes visible.
  useEffect(() => {
    const sentinel = sentinelRef.current;
    if (!sentinel || loading || !hasMore) return;

    const observer = new IntersectionObserver((entries) => {
      if (!entries[0].isIntersecting || loadingMoreRef.current) return;
      loadingMoreRef.current = true;
      setLoadingMore(true);
      void getLatestReports(query, locale, PAGE_SIZE, offsetRef.current)
        .then((page) => {
          offsetRef.current += page.reports.length;
          setHasMore(page.hasMore);
          setReports((current) => {
            const existingIds = new Set(current.map((report) => report.id));
            return [...current, ...page.reports.filter((report) => !existingIds.has(report.id))];
          });
        })
        .catch(() => setError(t("home.loadError")))
        .finally(() => {
          loadingMoreRef.current = false;
          setLoadingMore(false);
        });
    });
    observer.observe(sentinel);
    return () => observer.disconnect();
  }, [query, locale, t, loading, hasMore]);

  const toggleLike = async (reportId: number) => {
    try {
      const result = await toggleReportLike(reportId);
      setReports((items) =>
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

  return (
    <Stack spacing={4}>
      <Box
        sx={{
          display: "flex",
          justifyContent: "space-between",
          gap: 2,
          alignItems: { xs: "flex-start", sm: "center" },
          flexDirection: { xs: "column", sm: "row" },
        }}
      >
        <Box>
          <Typography
            variant="overline"
            color="secondary"
            sx={{ fontWeight: 700, letterSpacing: "0.14em" }}
          >
            {t("home.overline")}
          </Typography>
          <Typography
            variant="h2"
            sx={{ fontSize: { xs: "2.2rem", md: "3rem" } }}
          >
            {query ? t("home.title.search") : t("home.title.latest")}
          </Typography>
          <Typography color="text.secondary">{t("home.subtitle")}</Typography>
        </Box>
        <Button
          component={RouterLink}
          to="/submit"
          variant="contained"
          size="large"
          startIcon={<AddIcon />}
        >
          {t("home.addReport")}
        </Button>
      </Box>
      <TextField
        fullWidth
        value={query}
        onChange={(event) => setQuery(event.target.value)}
        placeholder={t("home.searchPlaceholder")}
        slotProps={{
          input: {
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            ),
          },
        }}
      />
      {error && <Alert severity="error">{error}</Alert>}
      {loading ? (
        <Box sx={{ display: "grid", placeItems: "center", py: 8 }}>
          <CircularProgress />
        </Box>
      ) : reports.length === 0 ? (
        <Alert severity="info">{t("home.empty")}</Alert>
      ) : (
        <Grid container spacing={2}>
          {reports.map((report) => (
            <Grid key={report.id} size={{ xs: 12, sm: 6, md: 4 }}>
              <Card variant="outlined" sx={{ height: "100%" }}>
                <CardActionArea
                  onClick={() => navigate(`/reports/${report.id}`)}
                >
                  <CardContent>
                    <Stack spacing={1.5}>
                      <Typography variant="h6">{report.product}</Typography>
                      <Typography color="text.secondary">
                        {report.brand}
                      </Typography>
                      <Typography
                        variant="body2"
                        sx={{
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                          display: "-webkit-box",
                          WebkitLineClamp: 3,
                          WebkitBoxOrient: "vertical",
                        }}
                      >
                        {report.description}
                      </Typography>
                      <Stack
                        direction="row"
                        spacing={1}
                        sx={{ flexWrap: "wrap" }}
                      >
                        <Chip label={report.category} size="small" />
                        <Chip
                          label={report.seller}
                          size="small"
                          variant="outlined"
                        />
                      </Stack>
                    </Stack>
                  </CardContent>
                </CardActionArea>
                <Box
                  sx={{ px: 2, pb: 1.5 }}
                  onClick={(event) => event.stopPropagation()}
                >
                  <LikeButton
                    liked={report.likedByMe}
                    count={report.likeCount}
                    onToggle={() => void toggleLike(report.id)}
                    label={t("reportDetail.like")}
                  />
                </Box>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}
      {!loading && hasMore && (
        <Box ref={sentinelRef} sx={{ display: "grid", placeItems: "center", py: 4 }}>
          {loadingMore && <CircularProgress size={28} />}
        </Box>
      )}
    </Stack>
  );
}
