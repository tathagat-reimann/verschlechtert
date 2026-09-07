import { useEffect, useState } from "react";
import {
  Alert,
  Button,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import SaveIcon from "@mui/icons-material/Save";
import AddPhotoAlternateIcon from "@mui/icons-material/AddPhotoAlternate";
import { useNavigate } from "react-router-dom";
import {
  createSubmission,
  getCatalogOptions,
  type CatalogOption,
  type CatalogOptions,
} from "../api";
import { storage } from "../firebase";
import { getDownloadURL, ref, uploadBytes } from "firebase/storage";
import { useLocale } from "../i18n/LocaleContext";
import { CatalogField } from "../components/CatalogField";

const MAX_IMAGES = 5;

export function NewSubmissionPage() {
  const navigate = useNavigate();
  const { t, locale } = useLocale();
  const [options, setOptions] = useState<CatalogOptions | null>(null);
  const [brand, setBrand] = useState<CatalogOption | null>(null);
  const [category, setCategory] = useState<CatalogOption | null>(null);
  const [seller, setSeller] = useState<CatalogOption | null>(null);
  const [form, setForm] = useState({
    productName: "",
    description: "",
    observedAt: "",
    productUrl: "",
  });
  const [images, setImages] = useState<File[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    void getCatalogOptions(locale)
      .then((next) => {
        setOptions(next);
        setCategory((current) =>
          current
            ? ([...next.categories, next.unspecifiedCategory].find(
                (option) => option.id === current.id,
              ) ?? current)
            : current,
        );
      })
      .catch(() => setError(t("newSubmission.catalogLoadError")));
  }, [locale, t]);

  const update = (key: string, value: string | number) =>
    setForm((current) => ({ ...current, [key]: value }));

  const chooseImages = (selected: FileList | null) => {
    if (!selected) return;
    const next = [...images, ...Array.from(selected)];
    if (next.length > MAX_IMAGES) {
      setError(t("newSubmission.maxImagesError"));
      return;
    }
    if (next.some((file) => !file.type.startsWith("image/"))) {
      setError(t("newSubmission.imageTypeError"));
      return;
    }
    setError(null);
    setImages(next);
  };

  const submit = async () => {
    if (!brand || !category || !seller) {
      setError(t("newSubmission.chooseFieldsError"));
      return;
    }
    setSaving(true);
    setError(null);
    try {
      const uploadedImages = await Promise.all(
        images.map(async (file, index) => {
          const storagePath = `reports/${crypto.randomUUID()}-${file.name}`;
          const snapshot = await uploadBytes(ref(storage, storagePath), file, {
            contentType: file.type,
          });
          return {
            storagePath,
            imageUrl: await getDownloadURL(snapshot.ref),
            sortOrder: index + 1,
          };
        }),
      );
      await createSubmission({
        ...form,
        brandId: brand.id,
        categoryId: category.id,
        sellerId: seller.id,
        images: uploadedImages,
      });
      navigate("/submissions");
    } catch {
      setError(t("newSubmission.saveError"));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Paper variant="outlined" sx={{ p: { xs: 3, md: 5 }, maxWidth: 760 }}>
      <Stack spacing={3}>
        <Typography
          variant="h2"
          sx={{ fontSize: { xs: "2.2rem", md: "3rem" } }}
        >
          {t("newSubmission.title")}
        </Typography>
        <Typography color="text.secondary">
          {t("newSubmission.subtitle")}
        </Typography>
        {error && <Alert severity="error">{error}</Alert>}
        <TextField
          label={t("newSubmission.productName")}
          required
          value={form.productName}
          onChange={(event) => update("productName", event.target.value)}
        />
        <CatalogField
          label={t("newSubmission.brand")}
          loadingLabel={t("newSubmission.brandLoading")}
          searchLabel={t("newSubmission.brandSearch")}
          notFoundSuffix={t("newSubmission.notFoundSuffix")}
          notFoundHint={t("newSubmission.notFoundHint")}
          options={options?.brands}
          notFoundOption={options?.unspecifiedBrand}
          value={brand}
          onChange={setBrand}
        />
        <CatalogField
          label={t("newSubmission.category")}
          loadingLabel={t("newSubmission.categoryLoading")}
          searchLabel={t("newSubmission.categorySearch")}
          notFoundSuffix={t("newSubmission.notFoundSuffix")}
          notFoundHint={t("newSubmission.notFoundHint")}
          options={options?.categories}
          notFoundOption={options?.unspecifiedCategory}
          value={category}
          onChange={setCategory}
        />
        <CatalogField
          label={t("newSubmission.seller")}
          loadingLabel={t("newSubmission.sellerLoading")}
          searchLabel={t("newSubmission.sellerSearch")}
          notFoundSuffix={t("newSubmission.notFoundSuffix")}
          notFoundHint={t("newSubmission.notFoundHint")}
          options={options?.sellers}
          notFoundOption={options?.unspecifiedSeller}
          value={seller}
          onChange={setSeller}
        />
        <TextField
          label={t("newSubmission.whatChanged")}
          required
          multiline
          minRows={4}
          value={form.description}
          onChange={(event) => update("description", event.target.value)}
        />
        <TextField
          label={t("newSubmission.observedAt")}
          type="date"
          slotProps={{ inputLabel: { shrink: true } }}
          value={form.observedAt}
          onChange={(event) => update("observedAt", event.target.value)}
        />
        <TextField
          label={t("newSubmission.productUrl")}
          value={form.productUrl}
          onChange={(event) => update("productUrl", event.target.value)}
        />
        <Stack spacing={1}>
          <Button
            component="label"
            variant="outlined"
            startIcon={<AddPhotoAlternateIcon />}
          >
            {t("newSubmission.addImages")} ({images.length}/{MAX_IMAGES})
            <input
              hidden
              type="file"
              accept="image/*"
              multiple
              onChange={(event) => {
                chooseImages(event.target.files);
                event.target.value = "";
              }}
            />
          </Button>
          {images.map((image, index) => (
            <Typography
              key={`${image.name}-${index}`}
              variant="body2"
              color="text.secondary"
            >
              {index + 1}. {image.name}
            </Typography>
          ))}
        </Stack>
        <Button
          variant="contained"
          size="large"
          startIcon={<SaveIcon />}
          onClick={() => void submit()}
          disabled={saving || !options}
        >
          {saving ? t("newSubmission.saving") : t("newSubmission.submit")}
        </Button>
      </Stack>
    </Paper>
  );
}
