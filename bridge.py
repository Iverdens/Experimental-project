import serial
import os

# 1. Настройка порта (проверь в диспетчере устройств, какой у тебя COM-порт)
ser = serial.Serial('COM4', 9600) 

print("System Bridge Active. Monitoring FPGA...")

# 2. Инициализируем файл безопасности
with open("security_status.txt", "w") as f:
    f.write("1")

while True:
    if ser.in_waiting > 0:
        data = ser.read(1)
        # Если ПЛИС прислала 0xFF (сигнал тревоги из твоего Verilog)
        if data == b'\xFF':
            print("ALERT: Hardware Interlock Triggered!")
            with open("security_status.txt", "w") as f:
                f.write("0")
        else:
            # Если всё в порядке
            with open("security_status.txt", "w") as f:
                f.write("1")