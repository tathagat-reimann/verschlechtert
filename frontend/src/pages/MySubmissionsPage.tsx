import { useEffect, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  Divider,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import SaveIcon from "@mui/icons-material/Save";
import { getCatalogOptions, getMySubmissions, toggleReportLike, updateSubmission, type CatalogOption, type CatalogOptions, type Submission } from "../api";
import { CatalogField } from "../components/CatalogField";
import { LikeButton } from "../components/LikeButton";
import { useLocale } from "../i18n/LocaleContext";
import { Link as RouterLink } from "react-router-dom";
import AddIcon from "@mui/icons-material/Add";

export function MySubmissionsPage() {
  const [submissions, setSubmissions] = useState<Submission[]>([]);
  const [options, setOptions] = useState<CatalogOptions | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editing, setEditing] = useState<number | null>(null);
  const [draft, setDraft] = useState({
    productName: "",
    description: "",
    observedAt: "",
    brand: null as CatalogOption | null,
    category: null as CatalogOption | null,
    seller: null as CatalogOption | null,
  });
  const { t, locale } = useLocale();

  useEffect(() => {
    void getMySubmissions(locale)
      .then(setSubmissions)
      .catch(() => setError(t("submissions.loadError")))
      .finally(() => setLoading(false));
  }, [locale, t]);

  useEffect(() => {
    void getCatalogOptions(locale).then(setOptions).catch(() => setError(t("newSubmission.catalogLoadError")));
  }, [locale, t]);

  const beginEdit = (submission: Submission) => {
    setEditing(submission.id);
    setDraft({
      productName: submission.product,
      description: submission.description,
      observedAt: submission.observedAt?.slice(0, 10) ?? "",
      brand: { id: submission.brandId, name: submission.brand },
      category: { id: submission.categoryId, name: submission.category },
      seller: { id: submission.sellerId, name: submission.seller },
    });
  };

  const save = async () => {
    if (editing === null) return;
    if (!draft.brand || !draft.category || !draft.seller) {
      setError(t("newSubmission.chooseFieldsError"));
      return;
    }
    try {
      await updateSubmission(editing, {
        productName: draft.productName,
        brandId: draft.brand.id,
        categoryId: draft.category.id,
        sellerId: draft.seller.id,
        description: draft.description,
        observedAt: draft.observedAt,
      });
      setSubmissions((items) =>
        items.map((item) =>
          item.id === editing
            ? {
                ...item,
                product: draft.productName,
                brand: draft.brand!.name,
                category: draft.category!.name,
                seller: draft.seller!.name,
                brandId: draft.brand!.id,
                categoryId: draft.category!.id,
                sellerId: draft.seller!.id,
                description: draft.description,
                observedAt: draft.observedAt,
              }
            : item,
        ),
      );
      setEditing(null);
    } catch {
      setError(t("submissions.updateError"));
    }
  };

  const toggleLike = async (reportId: number) => {
    try {
      const result = await toggleReportLike(reportId);
      setSubmissions((items) =>
        items.map((item) =>
          item.id === reportId ? { ...item, likedByMe: result.liked, likeCount: result.count } : item,
        ),
      );
    } catch {
      setError(t("reportDetail.likeError"));
    }
  };

  if (loading)
    return (
      <Stack sx={{ alignItems: "center", py: 8 }}>
        <CircularProgress />
      </Stack>
    );
  return (
    <Stack spacing={3}>
      <Box
        sx={{
          display: "flex",
          justifyContent: "space-between",
          gap: 2,
          alignItems: { xs: "flex-start", sm: "center" },
          flexDirection: { xs: "column", sm: "row" },
        }}
      >
        <BoxTitle />
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
      {error && <Alert severity="error">{error}</Alert>}
      {submissions.length === 0 ? (
        <Alert severity="info">{t("submissions.empty")}</Alert>
      ) : (
        submissions.map((submission) => (
          <Paper key={submission.id} variant="outlined" sx={{ p: 3 }}>
            <Stack spacing={2}>
              <Stack
                direction="row"
                sx={{ justifyContent: "space-between", gap: 2 }}
              >
                <Box>
                  <Typography variant="h5">{submission.product}</Typography>
                  <Typography color="text.secondary">
                    {submission.brand} · {submission.seller}
                  </Typography>
                </Box>
                <Chip
                  label={submission.status}
                  color={
                    submission.status === "approved" ? "success" : "default"
                  }
                />
              </Stack>
              <LikeButton
                liked={submission.likedByMe}
                count={submission.likeCount}
                onToggle={() => void toggleLike(submission.id)}
                label={t("reportDetail.like")}
              />
              <Divider />
              {editing === submission.id ? (
                <Stack spacing={2}>
                  <TextField
                    label={t("newSubmission.productName")}
                    value={draft.productName}
                    onChange={(e) =>
                      setDraft({ ...draft, productName: e.target.value })
                    }
                  />
                  <CatalogField
                    label={t("newSubmission.brand")}
                    loadingLabel={t("newSubmission.brandLoading")}
                    searchLabel={t("newSubmission.brandSearch")}
                    notFoundSuffix={t("newSubmission.notFoundSuffix")}
                    notFoundHint={t("newSubmission.notFoundHint")}
                    options={options?.brands}
                    notFoundOption={options?.unspecifiedBrand}
                    value={draft.brand}
                    onChange={(value) => setDraft({ ...draft, brand: value })}
                  />
                  <CatalogField
                    label={t("newSubmission.category")}
                    loadingLabel={t("newSubmission.categoryLoading")}
                    searchLabel={t("newSubmission.categorySearch")}
                    notFoundSuffix={t("newSubmission.notFoundSuffix")}
                    notFoundHint={t("newSubmission.notFoundHint")}
                    options={options?.categories}
                    notFoundOption={options?.unspecifiedCategory}
                    value={draft.category}
                    onChange={(value) => setDraft({ ...draft, category: value })}
                  />
                  <CatalogField
                    label={t("newSubmission.seller")}
                    loadingLabel={t("newSubmission.sellerLoading")}
                    searchLabel={t("newSubmission.sellerSearch")}
                    notFoundSuffix={t("newSubmission.notFoundSuffix")}
                    notFoundHint={t("newSubmission.notFoundHint")}
                    options={options?.sellers}
                    notFoundOption={options?.unspecifiedSeller}
                    value={draft.seller}
                    onChange={(value) => setDraft({ ...draft, seller: value })}
                  />
                  <TextField
                    label={t("submissions.description.label")}
                    multiline
                    minRows={3}
                    value={draft.description}
                    onChange={(e) =>
                      setDraft({ ...draft, description: e.target.value })
                    }
                  />
                  <TextField
                    label={t("submissions.observedAt.label")}
                    type="date"
                    slotProps={{ inputLabel: { shrink: true } }}
                    value={draft.observedAt}
                    onChange={(e) =>
                      setDraft({ ...draft, observedAt: e.target.value })
                    }
                  />
                  <Stack direction="row" spacing={1}>
                    <Button
                      variant="contained"
                      startIcon={<SaveIcon />}
                      onClick={() => void save()}
                      disabled={!options}
                    >
                      {t("submissions.save")}
                    </Button>
                    <Button onClick={() => setEditing(null)}>
                      {t("submissions.cancel")}
                    </Button>
                  </Stack>
                </Stack>
              ) : (
                <Stack spacing={1}>
                  <Typography>{submission.description}</Typography>
                  <Button
                    sx={{ alignSelf: "flex-start" }}
                    onClick={() => beginEdit(submission)}
                  >
                    {t("submissions.edit")}
                  </Button>
                </Stack>
              )}
            </Stack>
          </Paper>
        ))
      )}
    </Stack>
  );
}

function BoxTitle() {
  const { t } = useLocale();
  return (
    <Stack spacing={0.5}>
      <Typography
        variant="overline"
        color="secondary"
        sx={{ fontWeight: 700, letterSpacing: "0.14em" }}
      >
        {t("submissions.overline")}
      </Typography>
      <Typography variant="h2" sx={{ fontSize: { xs: "2.2rem", md: "3rem" } }}>
        {t("submissions.title")}
      </Typography>
      <Typography color="text.secondary">
        {t("submissions.subtitle")}
      </Typography>
    </Stack>
  );
}
