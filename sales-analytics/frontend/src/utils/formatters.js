
/**
 * Format currency values consistently throughout the application
 * @param {number} value - The number to format as currency
 * @param {string} locale - The locale to use for formatting (default: 'en-US')
 * @returns {string} Formatted currency string
 */
export const formatCurrency = (value, locale = 'en-US') => {
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    }).format(value);
  };
  
  /**
   * Format large numbers with comma separators
   * @param {number} value - The number to format
   * @param {string} locale - The locale to use for formatting (default: 'en-US')
   * @returns {string} Formatted number string
   */
  export const formatNumber = (value, locale = 'en-US') => {
    return new Intl.NumberFormat(locale).format(value);
  };
  
  /**
   * Format percentage values
   * @param {number} value - The number to format as percentage
   * @param {number} decimals - Number of decimal places (default: 1)
   * @returns {string} Formatted percentage string
   */
  export const formatPercentage = (value, decimals = 1) => {
    return `${value.toFixed(decimals)}%`;
  };
  
  /**
   * Format date/time consistently
   * @param {string|Date} date - The date to format
   * @param {string} format - The format to use (default: '24h')
   * @returns {string} Formatted time string
   */
  export const formatTime = (date, format = '24h') => {
    const d = new Date(date);
    if (format === '12h') {
      return d.toLocaleTimeString('en-US', { 
        hour: 'numeric', 
        minute: '2-digit', 
        hour12: true 
      });
    }
    return d.toLocaleTimeString('en-US', { 
      hour: '2-digit', 
      minute: '2-digit', 
      hour12: false 
    });
  };
  
  /**
   * Truncate long text with ellipsis
   * @param {string} text - The text to truncate
   * @param {number} length - Maximum length before truncation (default: 20)
   * @returns {string} Truncated text
   */
  export const truncateText = (text, length = 20) => {
    if (text.length <= length) return text;
    return `${text.substring(0, length)}...`;
  };
  
  /**
   * Calculate percentage change between two values
   * @param {number} current - Current value
   * @param {number} previous - Previous value
   * @returns {number} Percentage change
   */
  export const calculatePercentageChange = (current, previous) => {
    if (previous === 0) return 0;
    return ((current - previous) / previous) * 100;
  };