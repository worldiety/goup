//
//  ViewController.swift
//  iosApp
//
//  Created by Christian Maier on 08.05.19.
//  Copyright © 2019 Worldiety. All rights reserved.
//

import UIKit
import Myproject

class Callback : NSObject {
}

extension Callback : MyprojectPkgaHelloCallbackProtocol {
    func yourName() -> String {
        return "test"
    }
}


class ViewController: UIViewController {

    override func viewDidLoad() {
        super.viewDidLoad()
        
        print(MyprojectMyprojectAnExportedProjectLevelFunc())
        
        print(MyprojectPkgaNiceCallback(Callback()))
        
        
        var map = MyprojectPkgbGetMap2()
       // print(map!.get(map!.keys()!.get(1, error: nil)))
        let keys = map!.keys()!
        for i in 0..<keys.len() {
            let key = keys.get(i, error: nil)
            let key2 = keys.get(i, error: nil)

            let string = map!.get(key)
            let string2 = map!.get(key2)
            print("\(i) - \(key) - \(key2): \(string) - \(string2)")
            sleep(1)
        }
        
        
    }

}

