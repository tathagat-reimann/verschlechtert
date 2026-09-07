import { useEffect, useState } from "react";
import { Alert, Box, Button, Chip, CircularProgress, Divider, Paper, Stack, TextField, Typography } from "@mui/material";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import SendIcon from "@mui/icons-material/Send";
import { Link as RouterLink, useParams } from "react-router-dom";
import { addReportComment, getReportDetail, toggleReportLike, type ReportDetail } from "../api";
import { useLocale } from "../i18n/LocaleContext";
import { LikeButton } from "../components/LikeButton";

function formatDateTime(iso: string, locale: string) {
  const date = new Date(iso);
  const day = String(date.getDate()).padStart(2, "0");
  const month = date.toLocaleDateString(locale, { month: "short" });
  const time = date.toLocaleTimeString(locale, { hour: "2-digit", minute: "2-digit" });
  return `${day}.${month}.${date.getFullYear()} ${time}`;
}

export function ReportDetailPage() {
  const { id } = useParams();
  const reportId = Number(id);
  const { t, locale } = useLocale();
  const [report, setReport] = useState<ReportDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [commentBody, setCommentBody] = useState("");
  const [commentError, setCommentError] = useState<string | null>(null);
  const [submittingComment, setSubmittingComment] = useState(false);

  useEffect(() => {
    setLoading(true);
    void getReportDetail(reportId, locale).then(setReport).catch(() => setError(t("reportDetail.loadError"))).finally(() => setLoading(false));
  }, [reportId, locale, t]);

  const toggleLike = async () => {
    if (!report) return;
    try {
      const result = await toggleReportLike(report.id);
      setReport({ ...report, likedByMe: result.liked, likeCount: result.count });
    } catch {
      setError(t("reportDetail.likeError"));
    }
  };

  const submitComment = async () => {
    if (!report || !commentBody.trim()) return;
    setSubmittingComment(true);
    setCommentError(null);
    try {
      const comment = await addReportComment(report.id, commentBody);
      setReport({ ...report, comments: [...report.comments, comment] });
      setCommentBody("");
    } catch {
      setCommentError(t("reportDetail.commentError"));
    } finally {
      setSubmittingComment(false);
    }
  };

  if (loading) {
    return (
      <Stack sx={{ alignItems: "center", py: 8 }}>
        <CircularProgress />
      </Stack>
    );
  }
  if (error || !report) {
    return <Alert severity="error">{error ?? t("reportDetail.loadError")}</Alert>;
  }

  return (
    <Stack spacing={3}>
      <Button component={RouterLink} to="/" startIcon={<ArrowBackIcon />} sx={{ alignSelf: "flex-start" }}>
        {t("reportDetail.backToHome")}
      </Button>
      <Paper variant="outlined" sx={{ p: { xs: 3, md: 5 } }}>
        <Stack spacing={2}>
          <Stack direction="row" sx={{ justifyContent: "space-between", gap: 2, alignItems: "flex-start" }}>
            <Box>
              <Typography variant="h4">{report.product}</Typography>
              <Typography color="text.secondary">{report.brand} · {report.seller}</Typography>
            </Box>
            <Chip label={report.status} color={report.status === "approved" ? "success" : "default"} />
          </Stack>
          <Divider />
          <Typography>{report.description}</Typography>
          {report.observedAt && (
            <Typography variant="body2" color="text.secondary">{report.observedAt.slice(0, 10)}</Typography>
          )}
          {report.images.length > 0 && (
            <Stack direction="row" spacing={1} sx={{ flexWrap: "wrap" }}>
              {report.images.map((image) => (
                <Box key={image.imageUrl} component="img" src={image.imageUrl} sx={{ width: 160, height: 160, objectFit: "cover", borderRadius: 1 }} />
              ))}
            </Stack>
          )}
          <LikeButton liked={report.likedByMe} count={report.likeCount} onToggle={() => void toggleLike()} label={t("reportDetail.like")} />
        </Stack>
      </Paper>
      <Paper variant="outlined" sx={{ p: { xs: 3, md: 5 } }}>
        <Stack spacing={2}>
          <Typography variant="h6">{t("reportDetail.commentsTitle")}</Typography>
          {report.comments.length === 0 ? (
            <Typography color="text.secondary">{t("reportDetail.noComments")}</Typography>
          ) : (
            <Stack spacing={2}>
              {report.comments.map((comment) => (
                <Box key={comment.id}>
                  <Stack direction="row" spacing={1} sx={{ alignItems: "baseline" }}>
                    <Typography variant="subtitle2">{comment.authorName}</Typography>
                    <Typography variant="caption" color="text.secondary">
                      {formatDateTime(comment.createdAt, locale)}
                    </Typography>
                  </Stack>
                  <Typography variant="body2">{comment.body}</Typography>
                </Box>
              ))}
            </Stack>
          )}
          {report.isOwner ? (
            <Alert severity="info">{t("reportDetail.ownerHint")}</Alert>
          ) : (
            <Stack spacing={1}>
              <Divider />
              <TextField
                label={t("reportDetail.addCommentLabel")}
                placeholder={t("reportDetail.addCommentPlaceholder")}
                multiline
                minRows={2}
                value={commentBody}
                onChange={(event) => setCommentBody(event.target.value)}
              />
              {commentError && <Alert severity="error">{commentError}</Alert>}
              <Button
                variant="contained"
                startIcon={<SendIcon />}
                onClick={() => void submitComment()}
                disabled={submittingComment || !commentBody.trim()}
                sx={{ alignSelf: "flex-start" }}
              >
                {t("reportDetail.submitComment")}
              </Button>
            </Stack>
          )}
        </Stack>
      </Paper>
    </Stack>
  );
}
