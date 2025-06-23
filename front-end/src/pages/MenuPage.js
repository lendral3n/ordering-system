// src/pages/MenuPage.js
import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import {
  Box,
  Container,
  Grid,
  Typography,
  Tabs,
  Tab,
  TextField,
  InputAdornment,
  IconButton,
  Skeleton,
} from "@mui/material";
import { Search, FilterList } from "@mui/icons-material";
import MenuItemCard from "../components/Menu/MenuItemCard";
import MenuItemDialog from "../components/Menu/MenuItemDialog";
import FilterDialog from "../components/Menu/FilterDialog";
import { motion, AnimatePresence } from "framer-motion";
import {
  fetchCategories,
  fetchMenuItems,
  setSelectedCategory,
  setSearchQuery,
} from "../store/slices/menuSlice";

const MenuPage = () => {
  const dispatch = useDispatch();
  const {
    categories,
    items,
    selectedCategory,
    searchQuery,
    filters,
    isLoading,
  } = useSelector((state) => state.menu);

  const [selectedItem, setSelectedItem] = useState(null);
  const [showFilters, setShowFilters] = useState(false);

  useEffect(() => {
    dispatch(fetchCategories());
    dispatch(fetchMenuItems());
  }, [dispatch]);

  const handleCategoryChange = (event, newValue) => {
    dispatch(setSelectedCategory(newValue));
    dispatch(fetchMenuItems(newValue));
  };

  const handleSearch = (event) => {
    dispatch(setSearchQuery(event.target.value));
  };

  const filteredItems = items.filter((item) => {
    // Category filter
    if (selectedCategory && item.category_id !== selectedCategory) {
      return false;
    }

    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      return (
        item.name.toLowerCase().includes(query) ||
        (item.description && item.description.toLowerCase().includes(query))
      );
    }

    // Price filter
    if (
      item.price < filters.priceRange[0] ||
      item.price > filters.priceRange[1]
    ) {
      return false;
    }

    // Availability filter
    if (filters.available && !item.is_available) {
      return false;
    }

    return true;
  });

  return (
    <Container maxWidth="lg" sx={{ py: 2 }}>
      {/* Search Bar */}
      <Box sx={{ mb: 3 }}>
        <TextField
          fullWidth
          placeholder="Search menu items..."
          value={searchQuery}
          onChange={handleSearch}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <Search />
              </InputAdornment>
            ),
            endAdornment: (
              <InputAdornment position="end">
                <IconButton onClick={() => setShowFilters(true)}>
                  <FilterList />
                </IconButton>
              </InputAdornment>
            ),
          }}
        />
      </Box>

      {/* Categories */}
      <Box sx={{ borderBottom: 1, borderColor: "divider", mb: 3 }}>
        <Tabs
          value={selectedCategory || 0}
          onChange={handleCategoryChange}
          variant="scrollable"
          scrollButtons="auto">
          <Tab label="All" value={0} />
          {categories.map((category) => (
            <Tab key={category.id} label={category.name} value={category.id} />
          ))}
        </Tabs>
      </Box>

      {/* Menu Items Grid */}
      <Grid container spacing={2}>
        <AnimatePresence>
          {isLoading ? (
            // Loading skeletons
            [...Array(6)].map((_, index) => (
              <Grid item xs={12} sm={6} md={4} key={index}>
                <Skeleton variant="rectangular" height={300} />
              </Grid>
            ))
          ) : filteredItems.length === 0 ? (
            <Grid item xs={12}>
              <Box textAlign="center" py={5}>
                <Typography variant="h6" color="text.secondary">
                  No menu items found
                </Typography>
              </Box>
            </Grid>
          ) : (
            filteredItems.map((item) => (
              <Grid item xs={12} sm={6} md={4} key={item.id}>
                <motion.div
                  initial={{ opacity: 0, scale: 0.9 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.9 }}
                  transition={{ duration: 0.3 }}>
                  <MenuItemCard
                    item={item}
                    onClick={() => setSelectedItem(item)}
                  />
                </motion.div>
              </Grid>
            ))
          )}
        </AnimatePresence>
      </Grid>

      {/* Menu Item Dialog */}
      <MenuItemDialog
        item={selectedItem}
        open={!!selectedItem}
        onClose={() => setSelectedItem(null)}
      />

      {/* Filter Dialog */}
      <FilterDialog open={showFilters} onClose={() => setShowFilters(false)} />
    </Container>
  );
};

export default MenuPage;
