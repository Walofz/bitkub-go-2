const modeDisplay = document.getElementById('mode-display');
        const modeStatusBox = document.getElementById('mode-status-box');
        const lastRunDisplay = document.getElementById('last-run-display');
        const ethPriceDisplay = document.getElementById('eth-price-display');
        const totalValueDisplay = document.getElementById('total-value-display');
        const roiDisplay = document.getElementById('roi-display');
        const balanceTableBody = document.getElementById('balance-data');

        const numberFormatter = new Intl.NumberFormat('en-US', {
            minimumFractionDigits: 2,
            maximumFractionDigits: 2,
        });

        const coinFormatter = new Intl.NumberFormat('en-US', {
            minimumFractionDigits: 2, // อย่างน้อย 2 ตำแหน่งเพื่อความสม่ำเสมอ
            maximumFractionDigits: 8,
        });

        async function fetchStatus() {
            try {
                const response = await fetch('/api/status');
                const data = await response.json();

                modeDisplay.textContent = data.mode;
                modeStatusBox.className = 'status-box ' + (data.mode === 'DRY_RUN' ? 'dry-run' : 'production');
                lastRunDisplay.textContent = data.last_run;

                ethPriceDisplay.textContent = numberFormatter.format(data.coin_price || 0);
                totalValueDisplay.textContent = numberFormatter.format(data.total_value || 0) + ' THB';

                const roiValue = data.roi || 0;
                roiDisplay.textContent = roiValue.toFixed(2) + '%';
                roiDisplay.className = roiValue >= 0 ? 'roi-positive' : 'roi-negative';

                balanceTableBody.innerHTML = '';

                if (Array.isArray(data.portfolio)) {
                    data.portfolio.forEach(asset => {
                        const row = balanceTableBody.insertRow();

                        const displayCoinBalance = asset.coin_balance || 0;
                        const displayBalanceTHB = asset.balance_thb || 0;
                        const displayActualPct = asset.actual_pct || 0;
                        const displayTargetPct = asset.target_pct || 0;

                        const deviation = Math.abs(displayActualPct - displayTargetPct);
                        let rowClass = deviation > 5 ? 'style="background-color: #fff3cd;"' : '';
                        row.setAttribute('style', rowClass);

                        row.insertCell().textContent = asset.asset;

                        let coinBalanceText = '';
                        if (asset.asset === 'THB') {
                            coinBalanceText = numberFormatter.format(displayCoinBalance);
                        } else {
                            coinBalanceText = coinFormatter.format(displayCoinBalance);
                        }
                        row.insertCell().textContent = coinBalanceText;

                        row.insertCell().textContent = numberFormatter.format(displayBalanceTHB);
                        row.insertCell().textContent = displayActualPct.toFixed(2) + '%';
                        row.insertCell().textContent = displayTargetPct.toFixed(2) + '%';
                    });
                } else {
                    const row = balanceTableBody.insertRow();
                    row.insertCell(0).textContent = "ไม่พบข้อมูลพอร์ตโฟลิโอ หรือรูปแบบข้อมูลไม่ถูกต้อง";
                    row.cells[0].colSpan = 5;
                }

            } catch (error) {
                console.error('Error fetching status:', error);
                const row = balanceTableBody.insertRow();
                row.insertCell(0).textContent = "❌ ไม่สามารถเชื่อมต่อกับ Go Backend ได้";
                row.cells[0].colSpan = 5;
                balanceTableBody.innerHTML = row.outerHTML;
            }
        }

        async function fetchHistory() {
            try {
                const response = await fetch('/api/history');
                const data = await response.json();
                const tbody = document.getElementById('history-data');

                tbody.innerHTML = '';

                if (!data.trades || data.trades.length === 0) {
                    const row = tbody.insertRow();
                    row.innerHTML = `<td colspan="7" style="text-align: center;">ยังไม่มีประวัติการเทรดในโหมด Production</td>`;
                    return;
                }

                data.trades.forEach(trade => {
                    const row = tbody.insertRow();
                    row.insertCell().textContent = trade.timestamp;
                    row.insertCell().textContent = trade.asset;

                    const opCell = row.insertCell();
                    opCell.textContent = trade.operation.toUpperCase();
                    opCell.style.fontWeight = 'bold';
                    if (trade.operation === 'buy') {
                        opCell.style.color = 'green';
                    } else {
                        opCell.style.color = 'red';
                    }

                    row.insertCell().textContent = numberFormatter.format(trade.price);
                    row.insertCell().textContent = numberFormatter.format(trade.amount_thb);
                    row.insertCell().textContent = coinFormatter.format(trade.coin_amount);
                    row.insertCell().textContent = trade.deviation.toFixed(2) + '%';
                });

            } catch (error) {
                console.error('Error fetching history:', error);
            }
        }

        async function toggleMode(newMode) {
            if (confirm(`คุณแน่ใจหรือไม่ที่จะเปลี่ยนโหมดเป็น ${newMode.toUpperCase()}?`)) {
                try {
                    await fetch(`/api/mode/${newMode}`, { method: 'POST' });
                    fetchStatus();
                } catch (error) {
                    console.error('Error toggling mode:', error);
                    alert('เกิดข้อผิดพลาดในการเปลี่ยนโหมด');
                }
            }
        }

        setInterval(fetchStatus, 1000);
        setInterval(fetchHistory, 30000);
        fetchStatus();
        fetchHistory();