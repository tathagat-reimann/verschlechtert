import { IconButton, Stack, Typography } from "@mui/material";
import ThumbUpIcon from "@mui/icons-material/ThumbUp";
import ThumbUpOutlinedIcon from "@mui/icons-material/ThumbUpOutlined";

export function LikeButton({ liked, count, onToggle, label }: {
  liked: boolean;
  count: number;
  onToggle: () => void;
  label: string;
}) {
  return (
    <Stack direction="row" spacing={0.5} sx={{ alignItems: "center" }}>
      <IconButton size="small" color={liked ? "primary" : "default"} onClick={onToggle} aria-label={label}>
        {liked ? <ThumbUpIcon fontSize="small" /> : <ThumbUpOutlinedIcon fontSize="small" />}
      </IconButton>
      <Typography variant="body2" color="text.secondary">{count}</Typography>
    </Stack>
  );
}
