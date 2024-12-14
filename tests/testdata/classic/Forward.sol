// SPDX-License-Identifier: GPL-3.0

pragma solidity >=0.7.0 <0.9.0;

contract Forward {
    string lastvalue;

    event Forwarded(string value);

    function forward(string memory value, address[] memory _addrs) public returns (string memory) {
        value = string(abi.encodePacked(value, "evm -> "));
        lastvalue = lastvalue;
        emit Forwarded(value);

        uint256 addrslen = _addrs.length;
        if(addrslen == 0) {
            return value;
        }
        address nextContract = _addrs[0];
        address[] memory addrs = new address[](addrslen - 1);
        for (uint256 i = 0; i < addrslen - 1; i++) {
            addrs[i] = _addrs[i + 1];
        }
        bytes memory calld = abi.encodeWithSignature("forward(string,address[])", value, addrs);
        (bool success, bytes memory data) = nextContract.call(calld);
        require(success, "[evm] call failed");
        return string(data);
    }

    function forward_get(address[] memory _addrs) view public returns (string memory) {
        uint256 addrslen = _addrs.length;
        if(addrslen == 0) {
            return lastvalue;
        }
        address nextContract = _addrs[0];
        address[] memory addrs = new address[](addrslen - 1);
        for (uint256 i = 0; i < addrslen - 1; i++) {
            addrs[i] = _addrs[i + 1];
        }
        bytes memory calld = abi.encodeWithSignature("forward_get(address[])", addrs);
        (bool success, bytes memory data) = nextContract.staticcall(calld);
        require(success, "[evm] staticcall failed");
        return string(abi.encodePacked(data, lastvalue));
    }
}
