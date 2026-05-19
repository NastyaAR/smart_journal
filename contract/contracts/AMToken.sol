// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract AcademicMeritToken {
    string public name = "Academic Merit Token";
    string public symbol = "AMT";
    uint8 public decimals = 18;
    uint256 public totalSupply;

    mapping(address => uint256) public balanceOf;
    mapping(address => mapping(address => uint256)) public allowance;

    address public admin;

    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event Awarded(address indexed student, uint256 amount, string reason);
    event Redeemed(address indexed student, uint256 amount);

    modifier onlyAdmin() {
        require(msg.sender == admin, "Only admin");
        _;
    }

    constructor() {
        admin = msg.sender;
    }

    // ERC-20 стандартные методы
    function transfer(address to, uint256 value) public returns (bool) {
        require(balanceOf[msg.sender] >= value, "Insufficient balance");
        balanceOf[msg.sender] -= value;
        balanceOf[to] += value;
        emit Transfer(msg.sender, to, value);
        return true;
    }

    function approve(address spender, uint256 value) public returns (bool) {
        allowance[msg.sender][spender] = value;
        emit Approval(msg.sender, spender, value);
        return true;
    }

    function transferFrom(address from, address to, uint256 value) public returns (bool) {
        require(balanceOf[from] >= value, "Insufficient balance");
        require(allowance[from][msg.sender] >= value, "Not allowed");
        balanceOf[from] -= value;
        balanceOf[to] += value;
        allowance[from][msg.sender] -= value;
        emit Transfer(from, to, value);
        return true;
    }

    // Начисление токенов за достижения
    function awardStudent(address student, uint256 amount, string memory reason) public onlyAdmin {
        require(amount > 0, "Amount must be positive");
        uint256 mintAmount = amount * 10**decimals;
        balanceOf[student] += mintAmount;
        totalSupply += mintAmount;
        emit Transfer(address(0), student, mintAmount);
        emit Awarded(student, amount, reason);
    }

    // Обмен токенов на привилегии (сжигание)
    function redeem(uint256 amount) public {
        require(amount > 0, "Amount must be positive");
        uint256 burnAmount = amount * 10**decimals;
        require(balanceOf[msg.sender] >= burnAmount, "Insufficient balance");
        balanceOf[msg.sender] -= burnAmount;
        totalSupply -= burnAmount;
        emit Transfer(msg.sender, address(0), burnAmount);
        emit Redeemed(msg.sender, amount);
    }

    // Административное списание токенов студента при покупке мерча
    function redeemStudent(address student, uint256 amount) public onlyAdmin {
        require(amount > 0, "Amount must be positive");
        uint256 burnAmount = amount * 10**decimals;
        require(balanceOf[student] >= burnAmount, "Insufficient balance");
        balanceOf[student] -= burnAmount;
        totalSupply -= burnAmount;
        emit Transfer(student, address(0), burnAmount);
        emit Redeemed(student, amount);
    }

    // Получить баланс в AMT (без wei)
    function getBalance(address student) public view returns (uint256) {
        return balanceOf[student] / 10**decimals;
    }
}
