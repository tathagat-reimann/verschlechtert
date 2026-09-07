import { useState } from "react";
import { Alert, Autocomplete, Stack, TextField } from "@mui/material";
import { createFilterOptions } from "@mui/material/Autocomplete";
import type { CatalogOption } from "../api";

export const optionLabel = (option: CatalogOption) => option.name || option.slug || "";
const filterCatalogOptions = createFilterOptions<CatalogOption>();

export function CatalogField({ label, loadingLabel, searchLabel, notFoundSuffix, notFoundHint, options, notFoundOption, value, onChange }: {
  label: string;
  loadingLabel: string;
  searchLabel: string;
  notFoundSuffix: string;
  notFoundHint: string;
  options: CatalogOption[] | undefined;
  notFoundOption: CatalogOption | undefined;
  value: CatalogOption | null;
  onChange: (value: CatalogOption | null) => void;
}) {
  const [inputValue, setInputValue] = useState("");
  const isNotFound = notFoundOption != null && value?.id === notFoundOption.id;

  return (
    <Stack spacing={1}>
      <Autocomplete
        loading={!options}
        disabled={!options}
        options={options ?? []}
        value={value}
        inputValue={inputValue}
        onInputChange={(_, next) => setInputValue(next)}
        onChange={(_, selected) => onChange(selected)}
        getOptionLabel={optionLabel}
        isOptionEqualToValue={(option, val) => option.id === val.id}
        filterOptions={(opts, params) => {
          const filtered = filterCatalogOptions(opts, params);
          if (filtered.length === 0 && params.inputValue !== "" && notFoundOption) {
            return [notFoundOption];
          }
          return filtered;
        }}
        renderOption={(props, option) => (
          <li {...props} key={option.id}>
            {notFoundOption && option.id === notFoundOption.id ? `${option.name} ${notFoundSuffix}` : option.name}
          </li>
        )}
        renderInput={(params) => <TextField {...params} label={label} placeholder={!options ? loadingLabel : searchLabel} required />}
      />
      {isNotFound && <Alert severity="info">{notFoundHint}</Alert>}
    </Stack>
  );
}
