// src/pages/ScanPage.js
import React, { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import {
  Box,
  Container,
  Typography,
  TextField,
  Button,
  Paper,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  CircularProgress,
  Alert,
} from '@mui/material';
import { QrCodeScanner, Restaurant } from '@mui/icons-material';
import { QrReader } from 'react-qr-reader';
import { motion } from 'framer-motion';
import { startSession } from '../store/slices/sessionSlice';

const ScanPage = () => {
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { isLoading, error } = useSelector((state) => state.session);
  
  const [showScanner, setShowScanner] = useState(false);
  const [showForm, setShowForm] = useState(false);
  const [tableNumber, setTableNumber] = useState('');
  const [customerInfo, setCustomerInfo] = useState({
    name: '',
    phone: '',
  });

  const handleScan = useCallback((result) => {
    if (result) {
      const url = new URL(result.text);
      const table = url.searchParams.get('table');
      if (table) {
        setTableNumber(table);
        setShowScanner(false);
        setShowForm(true);
      }
    }
  }, []);

  const handleError = (error) => {
    console.error('QR Scanner error:', error);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    const result = await dispatch(startSession({
      tableNumber,
      customerName: customerInfo.name,
      customerPhone: customerInfo.phone,
    }));

    if (startSession.fulfilled.match(result)) {
      navigate('/menu');
    }
  };

  const handleManualEntry = () => {
    setShowForm(true);
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <Container maxWidth="sm">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
        >
          <Paper elevation={3} sx={{ p: 4, borderRadius: 2 }}>
            <Box textAlign="center" mb={4}>
              <Restaurant sx={{ fontSize: 60, color: 'primary.main', mb: 2 }} />
              <Typography variant="h4" gutterBottom fontWeight="bold">
                Welcome!
              </Typography>
              <Typography variant="body1" color="text.secondary">
                Scan the QR code on your table to start ordering
              </Typography>
            </Box>

            {error && (
              <Alert severity="error" sx={{ mb: 3 }}>
                {error}
              </Alert>
            )}

            {!showForm && (
              <Box>
                <Button
                  fullWidth
                  variant="contained"
                  size="large"
                  startIcon={<QrCodeScanner />}
                  onClick={() => setShowScanner(true)}
                  sx={{ mb: 2 }}
                >
                  Scan QR Code
                </Button>
                <Button
                  fullWidth
                  variant="outlined"
                  size="large"
                  onClick={handleManualEntry}
                >
                  Enter Table Number Manually
                </Button>
              </Box>
            )}

            {showForm && (
              <form onSubmit={handleSubmit}>
                <TextField
                  fullWidth
                  label="Table Number"
                  value={tableNumber}
                  onChange={(e) => setTableNumber(e.target.value)}
                  required
                  sx={{ mb: 2 }}
                />
                <TextField
                  fullWidth
                  label="Your Name"
                  value={customerInfo.name}
                  onChange={(e) => setCustomerInfo({ ...customerInfo, name: e.target.value })}
                  required
                  sx={{ mb: 2 }}
                />
                <TextField
                  fullWidth
                  label="Phone Number"
                  value={customerInfo.phone}
                  onChange={(e) => setCustomerInfo({ ...customerInfo, phone: e.target.value })}
                  sx={{ mb: 3 }}
                />
                <Button
                  fullWidth
                  type="submit"
                  variant="contained"
                  size="large"
                  disabled={isLoading}
                >
                  {isLoading ? <CircularProgress size={24} /> : 'Start Ordering'}
                </Button>
              </form>
            )}
          </Paper>
        </motion.div>
      </Container>

      {/* QR Scanner Dialog */}
      <Dialog
        open={showScanner}
        onClose={() => setShowScanner(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Scan QR Code</DialogTitle>
        <DialogContent>
          <Box sx={{ width: '100%', maxWidth: 500, mx: 'auto' }}>
            <QrReader
              onResult={handleScan}
              onError={handleError}
              constraints={{ facingMode: 'environment' }}
              style={{ width: '100%' }}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowScanner(false)}>Cancel</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default ScanPage;

