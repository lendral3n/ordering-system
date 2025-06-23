// src/components/Menu/FilterDialog.js
import React, { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Slider,
  Typography,
  Box,
  FormControlLabel,
  Switch,
} from "@mui/material";
import { setFilters } from "../../store/slices/menuSlice";
import { formatCurrency } from "../../utils/helpers";

const FilterDialog = ({ open, onClose }) => {
  const dispatch = useDispatch();
  const { filters } = useSelector((state) => state.menu);
  const [localFilters, setLocalFilters] = useState(filters);

  const handleApply = () => {
    dispatch(setFilters(localFilters));
    onClose();
  };

  const handleReset = () => {
    const defaultFilters = {
      priceRange: [0, 500000],
      available: true,
    };
    setLocalFilters(defaultFilters);
    dispatch(setFilters(defaultFilters));
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Filter Menu</DialogTitle>
      <DialogContent>
        <Box sx={{ pt: 2 }}>
          <Typography gutterBottom>Price Range</Typography>
          <Slider
            value={localFilters.priceRange}
            onChange={(e, newValue) =>
              setLocalFilters({ ...localFilters, priceRange: newValue })
            }
            valueLabelDisplay="auto"
            valueLabelFormat={(value) => formatCurrency(value)}
            min={0}
            max={500000}
            step={10000}
            sx={{ mb: 3 }}
          />
          <Box display="flex" justifyContent="space-between" mb={3}>
            <Typography variant="body2">
              {formatCurrency(localFilters.priceRange[0])}
            </Typography>
            <Typography variant="body2">
              {formatCurrency(localFilters.priceRange[1])}
            </Typography>
          </Box>

          <FormControlLabel
            control={
              <Switch
                checked={localFilters.available}
                onChange={(e) =>
                  setLocalFilters({
                    ...localFilters,
                    available: e.target.checked,
                  })
                }
              />
            }
            label="Available items only"
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleReset}>Reset</Button>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleApply} variant="contained">
          Apply
        </Button>
      </DialogActions>
    </Dialog>
  );
};

export default FilterDialog;
