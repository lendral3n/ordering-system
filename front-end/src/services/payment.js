// src/services/payment.js
export const initializeMidtrans = () => {
  const script = document.createElement("script");
  script.src = "https://app.sandbox.midtrans.com/snap/snap.js";
  script.setAttribute(
    "data-client-key",
    process.env.REACT_APP_MIDTRANS_CLIENT_KEY
  );
  document.head.appendChild(script);
};

export const payWithMidtrans = (token, options = {}) => {
  return new Promise((resolve, reject) => {
    if (!window.snap) {
      reject(new Error("Midtrans Snap not loaded"));
      return;
    }

    window.snap.pay(token, {
      onSuccess: (result) => {
        console.log("Payment success:", result);
        resolve(result);
      },
      onPending: (result) => {
        console.log("Payment pending:", result);
        resolve(result);
      },
      onError: (result) => {
        console.log("Payment error:", result);
        reject(result);
      },
      onClose: () => {
        console.log("Payment popup closed");
        reject(new Error("Payment cancelled"));
      },
      ...options,
    });
  });
};
